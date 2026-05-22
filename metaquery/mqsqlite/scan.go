// Package mqsqlite is the database/sql adapter for metaquery builders
// targeting SQLite. It executes a *metaquery.Builder and returns an untyped
// Result or a typed TypedResult[T] using only the standard library.
//
// The adapter is driver-agnostic: it works with any database/sql driver
// registered for SQLite (e.g. modernc.org/sqlite, mattn/go-sqlite3, libsql).
// The wrapped query's Dialect must be metaquery.DialectSQLite so the Builder
// emits SQLite-compatible ?N placeholders.
package mqsqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

// Queryer is the minimal database/sql surface the adapter needs. Implemented
// by *sql.DB, *sql.Tx, and *sql.Conn.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Run executes b and returns every row as a positional []any in the same
// order as Meta.Columns. If b.WantsTotal(), a follow-up COUNT(*) is issued
// and Meta.Pagination.Total is populated.
func Run(ctx context.Context, q Queryer, b *metaquery.Builder) (*metaquery.Result, error) {
	sqlText, args, err := b.Build()
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	data, err := collectValues(rows)
	if err != nil {
		return nil, err
	}

	meta := b.Meta()
	if err := populateTotal(ctx, q, b, &meta); err != nil {
		return nil, err
	}
	return &metaquery.Result{Data: data, Meta: meta}, nil
}

// Scan executes b, scans rows into []T using `db` struct tags (snake_case
// fallback), and returns a TypedResult with the same Meta envelope as Run.
// T is validated against b.OutputColumns() before querying.
func Scan[T any](ctx context.Context, q Queryer, b *metaquery.Builder) (*metaquery.TypedResult[T], error) {
	if err := metaquery.Validate[T](b); err != nil {
		return nil, err
	}
	sqlText, args, err := b.Build()
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data, err := collectStructs[T](rows)
	if err != nil {
		return nil, err
	}

	meta := b.Meta()
	if err := populateTotal(ctx, q, b, &meta); err != nil {
		return nil, err
	}
	return &metaquery.TypedResult[T]{Data: data, Meta: meta}, nil
}

// ScanOne runs b expecting exactly one row and returns the scanned T plus
// the Meta envelope.
func ScanOne[T any](ctx context.Context, q Queryer, b *metaquery.Builder) (T, metaquery.Meta, error) {
	var zero T
	if err := metaquery.Validate[T](b); err != nil {
		return zero, metaquery.Meta{}, err
	}
	sqlText, args, err := b.Build()
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	rows, err := q.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return zero, metaquery.Meta{}, err
		}
		return zero, metaquery.Meta{}, sql.ErrNoRows
	}
	v, err := scanRowInto[T](rows)
	if err != nil {
		return zero, metaquery.Meta{}, err
	}
	if err := rows.Err(); err != nil {
		return zero, metaquery.Meta{}, err
	}
	return v, b.Meta(), nil
}

func collectValues(rows *sql.Rows) ([][]any, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out [][]any
	for rows.Next() {
		buf := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range buf {
			ptrs[i] = &buf[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		out = append(out, buf)
	}
	return out, rows.Err()
}

func collectStructs[T any](rows *sql.Rows) ([]T, error) {
	var out []T
	for rows.Next() {
		v, err := scanRowInto[T](rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// scanRowInto scans the current row of rows into a freshly-allocated T. It
// resolves columns to struct fields by `db` tag (snake_case fallback) and
// fails fast on an unmapped column rather than silently dropping data.
func scanRowInto[T any](rows *sql.Rows) (T, error) {
	var zero T
	cols, err := rows.Columns()
	if err != nil {
		return zero, err
	}
	dst := reflect.New(reflect.TypeOf(zero)).Elem()
	fieldByName := buildFieldIndex(dst.Type())
	scanTargets := make([]any, len(cols))
	for i, c := range cols {
		idx, ok := fieldByName[c]
		if !ok {
			return zero, fmt.Errorf("mqsqlite: column %q has no matching field in %s", c, dst.Type().Name())
		}
		scanTargets[i] = dst.Field(idx).Addr().Interface()
	}
	if err := rows.Scan(scanTargets...); err != nil {
		return zero, err
	}
	return dst.Interface().(T), nil
}

// buildFieldIndex maps column names (from `db` tag or snake_case of the Go
// field name) to the field's index in t. Mirrors validate.go's name
// resolution so Scan and Validate agree on the mapping.
func buildFieldIndex(t reflect.Type) map[string]int {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	out := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := fieldColumnName(f)
		if name == "-" {
			continue
		}
		out[name] = i
	}
	return out
}

func fieldColumnName(f reflect.StructField) string {
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

func populateTotal(ctx context.Context, q Queryer, b *metaquery.Builder, meta *metaquery.Meta) error {
	if !b.WantsTotal() {
		return nil
	}
	sqlText, args, err := b.BuildCount()
	if err != nil {
		return err
	}
	var total int64
	if err := q.QueryRowContext(ctx, sqlText, args...).Scan(&total); err != nil {
		return err
	}
	if meta.Pagination == nil {
		meta.Pagination = &metaquery.Pagination{Total: -1}
	}
	meta.Pagination.Total = total
	return nil
}
