package metaquery

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode"
)

// Validate checks that T's exported fields map cleanly onto b.OutputColumns().
// A field matches a column by its `db` tag, falling back to snake_case of the
// field name. Drift in either direction (field with no column, column with no
// field) is an error.
//
// Called automatically by mqpgx.Scan[T]; call directly if you want to fail
// fast without issuing a query.
func Validate[T any](b *Builder) error {
	var zero T
	return validateShape(reflect.TypeOf(zero), b.OutputColumns())
}

func validateShape(t reflect.Type, cols []Column) error {
	if t == nil {
		return fmt.Errorf("metaquery: Validate[T] requires a concrete type")
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("metaquery: Validate[T] requires a struct, got %s", t.Kind())
	}

	colSet := make(map[string]struct{}, len(cols))
	for _, c := range cols {
		colSet[c.Name] = struct{}{}
		if c.OriginalName != "" && c.OriginalName != c.Name {
			colSet[c.OriginalName] = struct{}{}
		}
	}

	fieldNames := make(map[string]struct{})
	var fieldErrs []string
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := FieldColumnName(f)
		if name == "-" {
			continue
		}
		fieldNames[name] = struct{}{}
		if _, ok := colSet[name]; !ok {
			fieldErrs = append(fieldErrs, fmt.Sprintf("field %s.%s (column %q) not in projection", t.Name(), f.Name, name))
		}
	}

	var colErrs []string
	for _, c := range cols {
		if _, ok := fieldNames[c.Name]; ok {
			continue
		}
		if c.OriginalName != "" {
			if _, ok := fieldNames[c.OriginalName]; ok {
				continue
			}
		}
		colErrs = append(colErrs, fmt.Sprintf("column %q has no matching field in %s", c.Name, t.Name()))
	}

	if len(fieldErrs) == 0 && len(colErrs) == 0 {
		return nil
	}
	all := append(fieldErrs, colErrs...)
	return fmt.Errorf("metaquery: Validate[%s] shape mismatch: %s", t.Name(), strings.Join(all, "; "))
}

// FieldColumnName returns the column name a struct field should match
// against when scanning rows: the `db` tag if set, otherwise snake_case of
// the Go name. A `db:"-"` tag is returned verbatim so callers can skip the
// field.
//
// Exported so adapters (mqpgx, mqsqlite, future ones) and Validate share
// one source of truth for the field-to-column convention.
func FieldColumnName(f reflect.StructField) string {
	if tag, ok := f.Tag.Lookup("db"); ok {
		if comma := strings.IndexByte(tag, ','); comma >= 0 {
			tag = tag[:comma]
		}
		if tag != "" {
			return tag
		}
	}
	return toSnakeCase(f.Name)
}

func toSnakeCase(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 4)
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			sb.WriteByte('_')
		}
		sb.WriteRune(unicode.ToLower(r))
	}
	return sb.String()
}

// ValidateFilter checks that f is a sensible filter for q: the column exists,
// the op is permitted for the column's Go type, and the value is type-
// compatible (with JSON-friendly leniency — float64 for int columns, RFC3339
// strings for time columns).
//
// Raw-expression filters (f.IsExpr()) are passed through unchecked.
//
// Called automatically by Builder.ApplyFilter; call directly when you want
// to validate a filter you just deserialized from JSON without mutating a
// builder.
func ValidateFilter(q *Query, f Filter) error {
	if f.IsExpr() {
		return nil
	}
	if f.Column == "" {
		return fmt.Errorf("metaquery: filter missing column")
	}
	col := findColumn(q, f.Column)
	if col == nil {
		return fmt.Errorf("metaquery: unknown column %q", f.Column)
	}
	kind := valueKind(col.GoType)
	ops := allowedOps[kind]
	op := f.Op
	if op == "" {
		op = OpEq
	}
	if ops != nil && !slices.Contains(ops, op) {
		return fmt.Errorf("metaquery: op %q not valid for column %q (%s/%s)", op, col.Name, col.GoType, kind)
	}
	if op == OpIsNull || op == OpIsNotNull {
		return nil
	}
	return validateValue(kind, f.Value, col)
}

func findColumn(q *Query, name string) *Column {
	for i := range q.Columns {
		if q.Columns[i].Name == name || q.Columns[i].OriginalName == name {
			return &q.Columns[i]
		}
	}
	return nil
}

// valueKind maps a sqlc-emitted GoType string to a coarse kind used by both
// ValidateFilter and the plugin's column-kind emission.
func valueKind(goType string) string {
	switch goType {
	case "string", "pgtype.Text", "pgtype.Varchar", "pgtype.Name", "pgtype.BPChar":
		return "text"
	case "int", "int16", "int32", "int64", "uint", "uint16", "uint32", "uint64",
		"pgtype.Int2", "pgtype.Int4", "pgtype.Int8":
		return "int"
	case "float32", "float64", "pgtype.Float4", "pgtype.Float8":
		return "float"
	case "bool", "pgtype.Bool":
		return "bool"
	case "time.Time", "pgtype.Timestamp", "pgtype.Timestamptz", "pgtype.Date", "pgtype.Time":
		return "time"
	case "[]byte":
		return "bytes"
	default:
		return "any"
	}
}

var allowedOps = map[string][]Op{
	"text":  {OpEq, OpNe, OpLike, OpILike, OpIn, OpIsNull, OpIsNotNull},
	"int":   {OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpIn, OpIsNull, OpIsNotNull},
	"float": {OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpIsNull, OpIsNotNull},
	"bool":  {OpEq, OpNe, OpIsNull, OpIsNotNull},
	"time":  {OpEq, OpNe, OpLt, OpLe, OpGt, OpGe, OpIsNull, OpIsNotNull},
	"bytes": {OpEq, OpNe, OpIsNull, OpIsNotNull},
	// "any" intentionally absent — unknown kind skips op checking.
}

func validateValue(kind string, val any, col *Column) error {
	if val == nil {
		return nil
	}
	switch kind {
	case "text":
		if _, ok := val.(string); !ok {
			return valueTypeErr(col.Name, "string", val)
		}
	case "int":
		switch v := val.(type) {
		case int, int16, int32, int64, uint, uint16, uint32, uint64:
			return nil
		case float64:
			// JSON numbers come in as float64 — accept if losslessly integral.
			if v != float64(int64(v)) {
				return fmt.Errorf("metaquery: value %v for column %q not an integer", v, col.Name)
			}
			return nil
		default:
			return valueTypeErr(col.Name, "int", val)
		}
	case "float":
		switch val.(type) {
		case float32, float64, int, int16, int32, int64:
			return nil
		default:
			return valueTypeErr(col.Name, "float", val)
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return valueTypeErr(col.Name, "bool", val)
		}
	case "time":
		switch v := val.(type) {
		case time.Time:
			return nil
		case string:
			// JSON date strings (RFC3339) pass through; pgx parses for us.
			if _, err := time.Parse(time.RFC3339, v); err == nil {
				return nil
			}
			if _, err := time.Parse(time.RFC3339Nano, v); err == nil {
				return nil
			}
			return fmt.Errorf("metaquery: value %q for column %q not RFC3339 time", v, col.Name)
		default:
			return valueTypeErr(col.Name, "time.Time", val)
		}
	case "bytes":
		if _, ok := val.([]byte); !ok {
			if _, ok := val.(string); !ok {
				return valueTypeErr(col.Name, "[]byte", val)
			}
		}
	}
	return nil
}

func valueTypeErr(col, want string, got any) error {
	return fmt.Errorf("metaquery: value for column %q: want %s, got %T", col, want, got)
}
