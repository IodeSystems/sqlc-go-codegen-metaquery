package search

import (
	"strings"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

// Col is the column handle passed to a search Fn. Its methods build
// metaquery.Filters bound to the column, so a custom search reads like
// dataset's `search { f, s -> f.eq(...) }`. For virtual (Named) searches Col is
// the zero value (Column.Name == ""); use Expr to build a raw predicate.
type Col struct {
	Column  metaquery.Column
	dialect metaquery.Dialect
}

// Name is the column's whitelisted name.
func (c Col) Name() string { return c.Column.Name }

// Kind is the coarse type kind ("text"/"int"/"bool"/...) of the column.
func (c Col) Kind() string { return metaquery.ValueKind(c.Column.GoType) }

func (c Col) cmp(op metaquery.Op, v any) *metaquery.Filter {
	return &metaquery.Filter{Column: c.Column.Name, Op: op, Value: v}
}

// Comparison helpers. Each yields `<col> <op> ?` with v bound as a parameter.
func (c Col) Eq(v any) *metaquery.Filter { return c.cmp(metaquery.OpEq, v) }
func (c Col) Ne(v any) *metaquery.Filter { return c.cmp(metaquery.OpNe, v) }
func (c Col) Lt(v any) *metaquery.Filter { return c.cmp(metaquery.OpLt, v) }
func (c Col) Le(v any) *metaquery.Filter { return c.cmp(metaquery.OpLe, v) }
func (c Col) Gt(v any) *metaquery.Filter { return c.cmp(metaquery.OpGt, v) }
func (c Col) Ge(v any) *metaquery.Filter { return c.cmp(metaquery.OpGe, v) }

// IsNull / IsNotNull yield value-less predicates.
func (c Col) IsNull() *metaquery.Filter {
	return &metaquery.Filter{Column: c.Column.Name, Op: metaquery.OpIsNull}
}
func (c Col) IsNotNull() *metaquery.Filter {
	return &metaquery.Filter{Column: c.Column.Name, Op: metaquery.OpIsNotNull}
}

// Contains builds a case-insensitive substring match. Postgres uses ILIKE;
// SQLite uses LIKE (case-insensitive for ASCII by default). LIKE metacharacters
// in v are escaped so user input can't inject wildcards.
func (c Col) Contains(v string) *metaquery.Filter {
	op := "ILIKE"
	if c.dialect == metaquery.DialectSQLite {
		op = "LIKE"
	}
	return &metaquery.Filter{
		Expr: quoteIdent(c.Column.Name) + " " + op + ` ? ESCAPE '\'`,
		Args: []any{"%" + escapeLike(v) + "%"},
	}
}

// Expr builds a raw predicate with ? placeholders (renumbered at Build). This
// is the escape hatch for virtual searches and computed predicates; the caller
// owns its SQL safety, but values should still be passed as args, not inlined.
func (c Col) Expr(expr string, args ...any) *metaquery.Filter {
	return &metaquery.Filter{Expr: expr, Args: args}
}

// escapeLike escapes \ % _ so they're treated literally inside a LIKE/ILIKE
// pattern (paired with `ESCAPE '\'`).
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// quoteIdent double-quotes an identifier (valid for both Postgres and the
// SQLite builder, which also double-quotes).
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
