package golang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/internal/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// renderMetaQueries produces per-query metadata declarations for a single
// source file. Emission is gated by the emit_metaquery option, overridable
// per query with a `-- metaquery: <level>` comment.
//
// Ladder (additive): off < meta < wrap < cols.
//
// options is threaded in via a closure (see gen.go) so StructName uses the
// same initialisms/rename rules as the rest of the codegen.
func renderMetaQueries(queries []Query, source string, options *opts.Options) string {
	defaultLevel := levelValue(options.EmitMetaquery)
	var sb strings.Builder
	for _, q := range queries {
		if q.SourceName != source {
			continue
		}
		level := queryEmitLevel(q, defaultLevel)
		if level == emitOff {
			continue
		}
		renderMetaQuery(&sb, q)
		if level >= emitWrap && wrappableCmd(q.Cmd) {
			renderMetaWrapper(&sb, q)
		}
		if level >= emitCols {
			renderMetaCols(&sb, q, options)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// Emission levels form a numeric ladder so level >= N checks stay readable.
const (
	emitOff  = 0
	emitMeta = 1
	emitWrap = 2
	emitCols = 3
)

func levelValue(s string) int {
	switch s {
	case "off":
		return emitOff
	case "meta":
		return emitMeta
	case "wrap":
		return emitWrap
	case "cols", "":
		return emitCols
	}
	return emitCols
}

// queryEmitLevel scans q.Comments for `-- metaquery: <level>` and returns the
// resolved level (per-query override, else the provided default).
func queryEmitLevel(q Query, fallback int) int {
	for _, c := range q.Comments {
		s := strings.TrimSpace(c)
		s = strings.TrimPrefix(s, "--")
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "metaquery:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(s, "metaquery:"))
		return levelValue(val)
	}
	return fallback
}

// wrappableCmd reports whether a query Cmd makes sense to wrap in a Builder.
// :copyfrom has no placeholders; :batch* builds a separate batch object.
func wrappableCmd(cmd string) bool {
	switch cmd {
	case ":one", ":many", ":exec", ":execrows", ":execlastid", ":execresult":
		return true
	}
	return false
}

// renderMetaWrapper emits `func Wrap<Name>(...) *metaquery.Builder { ... }`
// with a typed parameter list derived from q.Arg.
func renderMetaWrapper(sb *strings.Builder, q Query) {
	params, callArgs := wrapperPieces(q)
	fmt.Fprintf(sb, "// Wrap%s returns a metaquery.Builder over Meta%s, pre-bound with typed arguments.\n", q.MethodName, q.MethodName)
	fmt.Fprintf(sb, "func Wrap%s(%s) *metaquery.Builder {\n", q.MethodName, params)
	if callArgs == "" {
		fmt.Fprintf(sb, "\treturn metaquery.Wrap(&Meta%s)\n", q.MethodName)
	} else {
		fmt.Fprintf(sb, "\treturn metaquery.Wrap(&Meta%s, %s)\n", q.MethodName, callArgs)
	}
	sb.WriteString("}\n")
}

// renderMetaCols emits a `var <Method>Cols = struct { ... }{ ... }` so callers
// get compile-time-safe column names and typed filter builders:
//
//	b.ApplyFilter(db.ListAuthorsCols.Name.ILike("%foo%"))
//	b.ApplyOrder(db.ListAuthorsCols.CreatedAt.Desc())
func renderMetaCols(sb *strings.Builder, q Query, options *opts.Options) {
	cols := metaColumns(q)
	if len(cols) == 0 {
		return
	}
	fmt.Fprintf(sb, "// %sCols gives typed, name-safe access to %s's output columns.\n", q.MethodName, q.MethodName)
	fmt.Fprintf(sb, "var %sCols = struct {\n", q.MethodName)
	fieldNames := make([]string, len(cols))
	seen := map[string]int{}
	for i, c := range cols {
		fn := colFieldName(c, options)
		if n, dup := seen[fn]; dup {
			seen[fn] = n + 1
			fn = fmt.Sprintf("%s_%d", fn, n+1)
		} else {
			seen[fn] = 1
		}
		fieldNames[i] = fn
		fmt.Fprintf(sb, "\t%s metaquery.%s\n", fn, columnKindFor(c.GoType))
	}
	sb.WriteString("}{\n")
	for i, c := range cols {
		fmt.Fprintf(sb, "\t%s: metaquery.%s(%q),\n", fieldNames[i], columnKindCtor(c.GoType), c.OriginalName)
	}
	sb.WriteString("}\n")
}

// columnKindFor maps a sqlc-emitted GoType to a metaquery col-kind type name.
// Unknowns fall back to AnyCol — the escape hatch with untyped ops.
func columnKindFor(goType string) string {
	switch goType {
	case "string", "pgtype.Text", "pgtype.Varchar", "pgtype.Name", "pgtype.BPChar":
		return "TextCol"
	case "int", "int16", "int32", "int64", "uint", "uint16", "uint32", "uint64",
		"pgtype.Int2", "pgtype.Int4", "pgtype.Int8":
		return "IntCol"
	case "float32", "float64", "pgtype.Float4", "pgtype.Float8":
		return "FloatCol"
	case "bool", "pgtype.Bool":
		return "BoolCol"
	case "time.Time", "pgtype.Timestamp", "pgtype.Timestamptz", "pgtype.Date", "pgtype.Time":
		return "TimeCol"
	case "[]byte":
		return "BytesCol"
	default:
		return "AnyCol"
	}
}

func columnKindCtor(goType string) string {
	return "New" + columnKindFor(goType)
}

// colFieldName is the Go field name used for a column inside the generated
// Cols struct. Uses sqlc's existing StructName so the field names align
// with the sqlc-generated model structs (ID, CreatedAt, AuthorID, ...).
func colFieldName(c metaCol, options *opts.Options) string {
	name := c.Name
	if name == "" {
		name = c.OriginalName
	}
	return StructName(name, options)
}

// wrapperPieces computes the Go parameter list for the wrapper function and
// the comma-separated expression list forwarded into metaquery.Wrap.
//
// Shapes:
//   - no args:               params="",                 callArgs=""
//   - single scalar arg:     params="id int64",         callArgs="id"
//   - multi-param (flat):    params="name string, bio pgtype.Text",
//                            callArgs="name, bio"
//   - struct param (emit):   params="arg CreateAuthorParams",
//                            callArgs="arg.Name, arg.Bio"
func wrapperPieces(q Query) (string, string) {
	if q.Arg.isEmpty() {
		return "", ""
	}
	pairs := q.Arg.Pairs()
	paramParts := make([]string, len(pairs))
	for i, p := range pairs {
		paramParts[i] = p.Name + " " + p.Type
	}
	params := strings.Join(paramParts, ", ")

	var callParts []string
	if q.Arg.Struct != nil {
		for _, f := range q.Arg.Struct.Fields {
			callParts = append(callParts, q.Arg.VariableForField(f))
		}
	} else {
		callParts = append(callParts, q.Arg.Name)
	}
	return params, strings.Join(callParts, ", ")
}

func renderMetaQuery(sb *strings.Builder, q Query) {
	fmt.Fprintf(sb, "var Meta%s = metaquery.Query{\n", q.MethodName)
	fmt.Fprintf(sb, "\tName: %q,\n", q.MethodName)
	fmt.Fprintf(sb, "\tCmd: %q,\n", q.Cmd)
	fmt.Fprintf(sb, "\tSource: %q,\n", q.SourceName)
	sb.WriteString("\tSQL: `")
	// backtick-escape: sqlc's SQL text may contain backticks in extension
	// function bodies; replace them with concatenation
	sb.WriteString(strings.ReplaceAll(q.SQL, "`", "`+\"`\"+`"))
	sb.WriteString("`,\n")

	cols := metaColumns(q)
	if len(cols) > 0 {
		sb.WriteString("\tColumns: []metaquery.Column{\n")
		for _, c := range cols {
			renderMetaColumn(sb, c)
		}
		sb.WriteString("\t},\n")
	}

	args := metaArgs(q)
	if len(args) > 0 {
		sb.WriteString("\tArgs: []metaquery.Arg{\n")
		for _, a := range args {
			renderMetaArg(sb, a)
		}
		sb.WriteString("\t},\n")
	}

	if q.Table != nil && (q.Table.Catalog != "" || q.Table.Schema != "" || q.Table.Name != "") {
		sb.WriteString("\tTable: &metaquery.Table{")
		if q.Table.Catalog != "" {
			fmt.Fprintf(sb, "Catalog: %q, ", q.Table.Catalog)
		}
		if q.Table.Schema != "" {
			fmt.Fprintf(sb, "Schema: %q, ", q.Table.Schema)
		}
		fmt.Fprintf(sb, "Name: %q", q.Table.Name)
		sb.WriteString("},\n")
	}

	sb.WriteString("}\n")
}

type metaCol struct {
	Name, OriginalName, GoType, DBType string
	Table, TableAlias                  string
	NotNull, IsArray                   bool
}

type metaArg struct {
	Position                int
	Name, GoType, DBType    string
	NotNull, IsArray, Slice bool
}

func metaColumns(q Query) []metaCol {
	if q.Ret.isEmpty() {
		return nil
	}
	if q.Ret.Struct != nil {
		out := make([]metaCol, 0, len(q.Ret.Struct.Fields))
		for _, f := range q.Ret.Struct.Fields {
			out = append(out, fieldToMetaCol(f))
		}
		return out
	}
	return []metaCol{scalarToMetaCol(q.Ret)}
}

func fieldToMetaCol(f Field) metaCol {
	c := metaCol{
		Name:         f.DBName,
		OriginalName: f.DBName,
		GoType:       f.Type,
	}
	if f.Column != nil {
		if f.Column.OriginalName != "" {
			c.OriginalName = f.Column.OriginalName
		}
		c.DBType = columnDBType(f.Column)
		c.NotNull = f.Column.NotNull
		c.IsArray = f.Column.IsArray
		if f.Column.Table != nil {
			c.Table = f.Column.Table.Name
		}
		c.TableAlias = f.Column.TableAlias
	}
	return c
}

func scalarToMetaCol(v QueryValue) metaCol {
	c := metaCol{
		Name:         v.DBName,
		OriginalName: v.DBName,
		GoType:       v.Typ,
	}
	if v.Column != nil {
		if v.Column.OriginalName != "" {
			c.OriginalName = v.Column.OriginalName
		}
		c.DBType = columnDBType(v.Column)
		c.NotNull = v.Column.NotNull
		c.IsArray = v.Column.IsArray
		if v.Column.Table != nil {
			c.Table = v.Column.Table.Name
		}
		c.TableAlias = v.Column.TableAlias
	}
	return c
}

func metaArgs(q Query) []metaArg {
	if q.Arg.isEmpty() {
		return nil
	}
	if q.Arg.Struct != nil {
		out := make([]metaArg, 0, len(q.Arg.Struct.Fields))
		for i, f := range q.Arg.Struct.Fields {
			a := metaArg{
				Position: i + 1,
				Name:     f.DBName,
				GoType:   f.Type,
			}
			if f.Column != nil {
				a.DBType = columnDBType(f.Column)
				a.NotNull = f.Column.NotNull
				a.IsArray = f.Column.IsArray
				a.Slice = f.Column.IsSqlcSlice
			}
			out = append(out, a)
		}
		return out
	}
	a := metaArg{
		Position: 1,
		Name:     q.Arg.DBName,
		GoType:   q.Arg.Typ,
	}
	if q.Arg.Column != nil {
		a.DBType = columnDBType(q.Arg.Column)
		a.NotNull = q.Arg.Column.NotNull
		a.IsArray = q.Arg.Column.IsArray
		a.Slice = q.Arg.Column.IsSqlcSlice
	}
	return []metaArg{a}
}

func columnDBType(c *plugin.Column) string {
	if c == nil || c.Type == nil {
		return ""
	}
	return c.Type.Name
}

func renderMetaColumn(sb *strings.Builder, c metaCol) {
	sb.WriteString("\t\t{")
	fmt.Fprintf(sb, "Name: %q, OriginalName: %q, GoType: %q", c.Name, c.OriginalName, c.GoType)
	if c.DBType != "" {
		fmt.Fprintf(sb, ", DBType: %q", c.DBType)
	}
	if c.NotNull {
		sb.WriteString(", NotNull: true")
	}
	if c.IsArray {
		sb.WriteString(", IsArray: true")
	}
	if c.Table != "" {
		fmt.Fprintf(sb, ", Table: %q", c.Table)
	}
	if c.TableAlias != "" {
		fmt.Fprintf(sb, ", TableAlias: %q", c.TableAlias)
	}
	sb.WriteString("},\n")
}

func renderMetaArg(sb *strings.Builder, a metaArg) {
	sb.WriteString("\t\t{")
	fmt.Fprintf(sb, "Position: %s", strconv.Itoa(a.Position))
	if a.Name != "" {
		fmt.Fprintf(sb, ", Name: %q", a.Name)
	}
	fmt.Fprintf(sb, ", GoType: %q", a.GoType)
	if a.DBType != "" {
		fmt.Fprintf(sb, ", DBType: %q", a.DBType)
	}
	if a.NotNull {
		sb.WriteString(", NotNull: true")
	}
	if a.IsArray {
		sb.WriteString(", IsArray: true")
	}
	if a.Slice {
		sb.WriteString(", IsSqlcSlice: true")
	}
	sb.WriteString("},\n")
}
