package search

import "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"

// EnumFn builds an exact-match Fn for an enum (or other specialized text-like)
// column. On Postgres it compares the column cast to text — `col::text = $` —
// rather than casting the user's value to the enum type (`$::mood`), because an
// invalid enum input would raise a query error instead of simply not matching.
// On SQLite (no enum types) it is a plain `col = $`.
//
// It is the type-driven default for text-kind columns whose DBType isn't a
// plain text type; assign it explicitly via Config.Fields for enum columns that
// sqlc emits as a custom Go type (which read as the "any" kind).
func EnumFn() Fn {
	return func(c Col, v string) (*metaquery.Filter, error) {
		if c.dialect == metaquery.DialectSQLite {
			return c.Eq(v), nil
		}
		return &metaquery.Filter{
			Expr: quoteIdent(c.Column.Name) + "::text = ?",
			Args: []any{v},
		}, nil
	}
}
