package metaquery

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Direction for OrderBy.
type Direction string

const (
	Asc  Direction = "ASC"
	Desc Direction = "DESC"
)

// Builder composes filters, projection, grouping, aggregations, ordering, and
// pagination on top of a Query. Build returns the final SQL wrapping the
// original query as a CTE, plus the full positional argument slice.
//
// Zero value is not usable; construct with Wrap.
type Builder struct {
	q       *Query
	base    []any
	selects []string
	wheres  []Filter
	groupBy []groupByEntry
	having  []Filter
	aggs    []aggregate
	orderBy []OrderBy
	limit   *int
	offset  *int
	total   bool
	errs    []error
}

type aggregate struct {
	alias  string
	expr   string
	goType string // output Go type of the aggregate
	dbType string // output DB type (best-effort)
}

type groupByEntry struct {
	col    string // whitelisted column name (quoted in output)
	expr   string // raw SQL expression with ? placeholders; mutually exclusive with col
	args   []any  // bound args for expr placeholders
	alias  string // output alias for expr-based entries
	goType string // Go type for expr-based entries (used in OutputColumns)
	dbType string // DB type for expr-based entries
}

// Wrap creates a Builder over q. baseArgs are the positional arguments for the
// original query ($1..$N); the Builder's own added arguments renumber to
// $N+1 and up.
func Wrap(q *Query, baseArgs ...any) *Builder {
	return &Builder{
		q:    q,
		base: append([]any(nil), baseArgs...),
	}
}

// ---- WHERE ----

// Where appends `<col> <op> ?`, whitelisting col against q.Columns. Prefer
// passing a typed Op constant (OpEq, OpILike, ...) — raw string literals
// still work because they're untyped constants.
func (b *Builder) Where(col string, op Op, val any) *Builder {
	return b.ApplyFilter(Filter{Column: col, Op: op, Value: val})
}

// WhereExpr appends a raw predicate expression using ? placeholders. Caller
// owns the safety of expr.
func (b *Builder) WhereExpr(expr string, args ...any) *Builder {
	return b.ApplyFilter(Filter{Expr: expr, Args: args})
}

// ApplyFilter appends a Filter to the WHERE clause. See Filter for the two
// valid shapes. Validates via ValidateFilter; accumulated errors surface at
// Build time.
func (b *Builder) ApplyFilter(f Filter) *Builder {
	if err := ValidateFilter(b.q, f); err != nil {
		b.errs = append(b.errs, err)
		return b
	}
	b.wheres = append(b.wheres, f)
	return b
}

// ApplyFilters is a convenience for applying a slice of filters.
func (b *Builder) ApplyFilters(fs []Filter) *Builder {
	for _, f := range fs {
		b.ApplyFilter(f)
	}
	return b
}

// ---- HAVING ----

// Having appends a raw HAVING predicate.
func (b *Builder) Having(expr string, args ...any) *Builder {
	return b.ApplyHaving(Filter{Expr: expr, Args: args})
}

// ApplyHaving appends a Filter to the HAVING clause. Aggregate expressions
// are common in HAVING so raw-expr filters are typical here; structured
// filters go through the same ValidateFilter check as WHERE.
func (b *Builder) ApplyHaving(f Filter) *Builder {
	if err := ValidateFilter(b.q, f); err != nil {
		b.errs = append(b.errs, err)
		return b
	}
	b.having = append(b.having, f)
	return b
}

// ---- ORDER BY / PAGINATION ----

// OrderBy appends an ORDER BY clause, whitelisting col against q.Columns.
func (b *Builder) OrderBy(col string, dir Direction) *Builder {
	return b.ApplyOrder(OrderBy{Column: col, Dir: dir})
}

// OrderByExpr appends a raw ORDER BY fragment (no whitelist).
func (b *Builder) OrderByExpr(expr string) *Builder {
	return b.ApplyOrder(OrderBy{Expr: expr})
}

// ApplyOrder appends an OrderBy. Exactly one of Column/Expr must be set.
func (b *Builder) ApplyOrder(o OrderBy) *Builder {
	if o.Column == "" && o.Expr == "" {
		b.errs = append(b.errs, errors.New("metaquery: OrderBy needs Column or Expr"))
		return b
	}
	if o.Column != "" && !b.q.HasColumn(o.Column) {
		b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in OrderBy", o.Column))
		return b
	}
	if o.Column != "" && o.Dir == "" {
		o.Dir = Asc
	}
	b.orderBy = append(b.orderBy, o)
	return b
}

// ApplyOrders is a convenience for a slice of OrderBy entries.
func (b *Builder) ApplyOrders(os []OrderBy) *Builder {
	for _, o := range os {
		b.ApplyOrder(o)
	}
	return b
}

// Limit sets the LIMIT value.
func (b *Builder) Limit(n int) *Builder {
	b.limit = &n
	return b
}

// Offset sets the OFFSET value.
func (b *Builder) Offset(n int) *Builder {
	b.offset = &n
	return b
}

// WithTotal opts into running a second COUNT(*) query when the builder is
// executed; result populates Meta.Pagination.Total.
func (b *Builder) WithTotal() *Builder {
	b.total = true
	return b
}

// Paginate is a convenience for Limit + Offset (and WithTotal when
// req.Total is set). Page is 0-indexed.
func (b *Builder) Paginate(page, size int) *Builder {
	if size > 0 {
		b.limit = &size
		if page > 0 {
			off := page * size
			b.offset = &off
		}
	}
	return b
}

// ApplyPagination drives pagination from a structured request.
func (b *Builder) ApplyPagination(r PageRequest) *Builder {
	b.Paginate(r.Page, r.Size)
	if r.Total {
		b.total = true
	}
	return b
}

// ---- SELECT / GROUP BY / AGG ----

// Select restricts the projection to the given columns, whitelisted. Without
// a Select (or GroupBy/Agg), projection defaults to *.
func (b *Builder) Select(cols ...string) *Builder {
	for _, c := range cols {
		if !b.q.HasColumn(c) {
			b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in Select", c))
			continue
		}
		b.selects = append(b.selects, c)
	}
	return b
}

// GroupBy appends GROUP BY columns, whitelisted.
func (b *Builder) GroupBy(cols ...string) *Builder {
	for _, c := range cols {
		if !b.q.HasColumn(c) {
			b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in GroupBy", c))
			continue
		}
		b.groupBy = append(b.groupBy, groupByEntry{col: c})
	}
	return b
}

// GroupByExpr appends a raw SQL expression to the GROUP BY clause. The result
// is projected as `<expr> AS <alias>` in the SELECT and appears in
// OutputColumns with the given goType. Use ? placeholders for bound args.
// This is the escape hatch for grouping by computed expressions (e.g. JSON
// field extraction: `GroupByExpr("group_value", "context_json->>?", "string", key)`).
func (b *Builder) GroupByExpr(alias, expr, goType string, args ...any) *Builder {
	if alias == "" || expr == "" {
		b.errs = append(b.errs, errors.New("metaquery: GroupByExpr requires non-empty alias and expr"))
		return b
	}
	b.groupBy = append(b.groupBy, groupByEntry{expr: expr, args: args, alias: alias, goType: goType, dbType: "text"})
	return b
}

// Agg adds an aggregate to the projection. alias is the output column name,
// expr is raw SQL, goType is the Go type of the result (used when building
// Meta.Columns and validating Into[T] shapes).
func (b *Builder) Agg(alias, expr, goType string) *Builder {
	if alias == "" || expr == "" {
		b.errs = append(b.errs, errors.New("metaquery: Agg requires non-empty alias and expr"))
		return b
	}
	b.aggs = append(b.aggs, aggregate{alias: alias, expr: expr, goType: goType})
	return b
}

// Count adds `count(*) AS <alias>` projecting int64 NOT NULL.
func (b *Builder) Count(alias string) *Builder {
	return b.addAgg(alias, "count(*)", "int64", "bigint", true)
}

// CountOf adds `count(<col>) AS <alias>` projecting int64 NOT NULL. col is
// whitelisted.
func (b *Builder) CountOf(alias, col string) *Builder {
	if !b.q.HasColumn(col) {
		b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in CountOf", col))
		return b
	}
	return b.addAgg(alias, "count("+quoteIdent(col)+")", "int64", "bigint", true)
}

// Sum, Avg, Min, Max project the same GoType as the source column but
// nullable (aggregates over empty groups return NULL). Use the Agg escape
// hatch with coalesce(...) to force NOT NULL.
func (b *Builder) Sum(alias, col string) *Builder {
	return b.columnAgg(alias, "sum", col)
}

func (b *Builder) Avg(alias, col string) *Builder {
	c := b.lookupColumn(col)
	if c == nil {
		b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in Avg", col))
		return b
	}
	// avg() returns numeric for integer inputs in pg; callers usually want float64.
	return b.addAgg(alias, "avg("+quoteIdent(col)+")", "float64", "numeric", false)
}

func (b *Builder) Min(alias, col string) *Builder {
	return b.columnAgg(alias, "min", col)
}

func (b *Builder) Max(alias, col string) *Builder {
	return b.columnAgg(alias, "max", col)
}

func (b *Builder) columnAgg(alias, fn, col string) *Builder {
	c := b.lookupColumn(col)
	if c == nil {
		b.errs = append(b.errs, fmt.Errorf("metaquery: unknown column %q in %s", col, fn))
		return b
	}
	return b.addAgg(alias, fn+"("+quoteIdent(col)+")", c.GoType, c.DBType, false)
}

func (b *Builder) lookupColumn(name string) *Column {
	for i := range b.q.Columns {
		if b.q.Columns[i].Name == name || b.q.Columns[i].OriginalName == name {
			return &b.q.Columns[i]
		}
	}
	return nil
}

func (b *Builder) groupByNames() []string {
	out := make([]string, len(b.groupBy))
	for i, g := range b.groupBy {
		if g.expr != "" {
			out[i] = g.alias
		} else {
			out[i] = g.col
		}
	}
	return out
}

func (b *Builder) addAgg(alias, expr, goType, dbType string, notNull bool) *Builder {
	_ = notNull
	b.aggs = append(b.aggs, aggregate{alias: alias, expr: expr, goType: goType, dbType: dbType})
	return b
}

// ---- BUILD ----

// OutputColumns returns the ordered set of columns the builder's final SELECT
// will produce. Used for Meta and for Into[T] validation.
func (b *Builder) OutputColumns() []Column {
	if len(b.groupBy) > 0 || len(b.aggs) > 0 {
		out := make([]Column, 0, len(b.groupBy)+len(b.aggs))
		for _, g := range b.groupBy {
			if g.expr != "" {
				out = append(out, Column{Name: g.alias, GoType: g.goType, DBType: g.dbType})
			} else if c := b.lookupColumn(g.col); c != nil {
				out = append(out, *c)
			} else {
				out = append(out, Column{Name: g.col})
			}
		}
		for _, a := range b.aggs {
			out = append(out, Column{Name: a.alias, GoType: a.goType, DBType: a.dbType, NotNull: false})
		}
		return out
	}
	if len(b.selects) > 0 {
		out := make([]Column, 0, len(b.selects))
		for _, s := range b.selects {
			if c := b.lookupColumn(s); c != nil {
				out = append(out, *c)
			} else {
				out = append(out, Column{Name: s})
			}
		}
		return out
	}
	return append([]Column(nil), b.q.Columns...)
}

// Meta returns the envelope describing the builder's applied state. Meta does
// NOT include Pagination.Total — that's populated by the adapter after the
// count query runs.
func (b *Builder) Meta() Meta {
	m := Meta{
		Columns: b.OutputColumns(),
		Where:   append([]Filter(nil), b.wheres...),
		Having:  append([]Filter(nil), b.having...),
		GroupBy: b.groupByNames(),
		OrderBy: append([]OrderBy(nil), b.orderBy...),
	}
	if b.limit != nil || b.offset != nil || b.total {
		p := &Pagination{Total: -1}
		if b.limit != nil {
			p.Limit = *b.limit
		}
		if b.offset != nil {
			p.Offset = *b.offset
		}
		m.Pagination = p
	}
	return m
}

// WantsTotal reports whether the caller requested a total-row count (see
// WithTotal / ApplyPagination). Adapters use this to run a follow-up COUNT.
func (b *Builder) WantsTotal() bool { return b.total }

// Build assembles the final SQL and argument slice.
func (b *Builder) Build() (string, []any, error) {
	if len(b.errs) > 0 {
		return "", nil, errors.Join(b.errs...)
	}

	var sb strings.Builder
	sb.WriteString("WITH __q AS (\n")
	sb.WriteString(b.q.SQL)
	sb.WriteString("\n)\nSELECT ")

	args := append([]any(nil), b.base...)
	nextPos := len(args) + 1
	dialect := b.q.Dialect

	proj, projArgs, nextPos := b.buildProjection(nextPos, dialect)
	sb.WriteString(proj)
	args = append(args, projArgs...)
	sb.WriteString(" FROM __q")

	if len(b.wheres) > 0 {
		sb.WriteString(" WHERE ")
		for i, f := range b.wheres {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			expr, fargs, np := renderFilter(f, nextPos, dialect)
			sb.WriteString(expr)
			args = append(args, fargs...)
			nextPos = np
		}
	}

	if len(b.groupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		for i, g := range b.groupBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			if g.expr != "" {
				sb.WriteString(quoteIdent(g.alias))
			} else {
				sb.WriteString(quoteIdent(g.col))
			}
		}
	}

	if len(b.having) > 0 {
		sb.WriteString(" HAVING ")
		for i, f := range b.having {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			expr, fargs, np := renderFilter(f, nextPos, dialect)
			sb.WriteString(expr)
			args = append(args, fargs...)
			nextPos = np
		}
	}

	if len(b.orderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, o := range b.orderBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			if o.Expr != "" {
				sb.WriteString(o.Expr)
				continue
			}
			sb.WriteString(quoteIdent(o.Column))
			sb.WriteByte(' ')
			sb.WriteString(string(o.Dir))
		}
	}

	if b.limit != nil {
		sb.WriteString(" LIMIT ")
		sb.WriteString(strconv.Itoa(*b.limit))
	}
	if b.offset != nil {
		sb.WriteString(" OFFSET ")
		sb.WriteString(strconv.Itoa(*b.offset))
	}

	return sb.String(), args, nil
}

// BuildCount returns a SELECT count(*) query over the same CTE + WHERE
// filters (but without GROUP BY / ORDER BY / LIMIT / OFFSET). Aggregation
// builders (groupBy non-empty) wrap the whole projection so distinct groups
// are counted.
func (b *Builder) BuildCount() (string, []any, error) {
	if len(b.errs) > 0 {
		return "", nil, errors.Join(b.errs...)
	}

	var sb strings.Builder
	sb.WriteString("WITH __q AS (\n")
	sb.WriteString(b.q.SQL)
	sb.WriteString("\n)")

	args := append([]any(nil), b.base...)
	nextPos := len(args) + 1
	dialect := b.q.Dialect

	var filtered strings.Builder
	filtered.WriteString(" FROM __q")
	if len(b.wheres) > 0 {
		filtered.WriteString(" WHERE ")
		for i, f := range b.wheres {
			if i > 0 {
				filtered.WriteString(" AND ")
			}
			expr, fargs, np := renderFilter(f, nextPos, dialect)
			filtered.WriteString(expr)
			args = append(args, fargs...)
			nextPos = np
		}
	}

	if len(b.groupBy) > 0 {
		// Count distinct groups: wrap the grouped projection in a subquery.
		sb.WriteString("\nSELECT count(*) FROM (SELECT ")
		proj, projArgs, _ := b.buildProjection(nextPos, dialect)
		sb.WriteString(proj)
		args = append(args, projArgs...)
		sb.WriteString(filtered.String())
		sb.WriteString(" GROUP BY ")
		for i, g := range b.groupBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			if g.expr != "" {
				sb.WriteString(quoteIdent(g.alias))
			} else {
				sb.WriteString(quoteIdent(g.col))
			}
		}
		if len(b.having) > 0 {
			sb.WriteString(" HAVING ")
			for i, f := range b.having {
				if i > 0 {
					sb.WriteString(" AND ")
				}
				expr, fargs, np := renderFilter(f, nextPos, dialect)
				sb.WriteString(expr)
				args = append(args, fargs...)
				nextPos = np
			}
		}
		sb.WriteString(") __c")
		return sb.String(), args, nil
	}

	sb.WriteString("\nSELECT count(*)")
	sb.WriteString(filtered.String())
	return sb.String(), args, nil
}

func (b *Builder) buildProjection(startPos int, dialect Dialect) (string, []any, int) {
	if len(b.groupBy) > 0 || len(b.aggs) > 0 {
		parts := make([]string, 0, len(b.groupBy)+len(b.aggs))
		var args []any
		pos := startPos
		for _, g := range b.groupBy {
			if g.expr != "" {
				expr, gargs, np := renumber(g.expr, g.args, pos, dialect)
				parts = append(parts, expr+" AS "+quoteIdent(g.alias))
				args = append(args, gargs...)
				pos = np
			} else {
				parts = append(parts, quoteIdent(g.col))
			}
		}
		for _, a := range b.aggs {
			parts = append(parts, a.expr+" AS "+quoteIdent(a.alias))
		}
		return strings.Join(parts, ", "), args, pos
	}
	if len(b.selects) > 0 {
		parts := make([]string, len(b.selects))
		for i, c := range b.selects {
			parts[i] = quoteIdent(c)
		}
		return strings.Join(parts, ", "), nil, startPos
	}
	return "*", nil, startPos
}

// renderFilter produces a `col op <ph>` or raw expr with ? renumbered to the
// next free placeholder slot. IS NULL / IS NOT NULL are value-less and emit
// no placeholder. The placeholder syntax (`$N` vs `?N`) is selected by
// dialect. ILIKE is translated to LIKE for SQLite (no ILIKE; LIKE is
// case-insensitive for ASCII by default).
func renderFilter(f Filter, start int, dialect Dialect) (string, []any, int) {
	if f.Expr != "" {
		return renumber(f.Expr, f.Args, start, dialect)
	}
	op := f.Op
	if op == "" {
		op = OpEq
	}
	if dialect == DialectSQLite && op == OpILike {
		op = OpLike
	}
	if op == OpIsNull || op == OpIsNotNull {
		return quoteIdent(f.Column) + " " + string(op), nil, start
	}
	expr, args, np := renumber("?", []any{f.Value}, start, dialect)
	return quoteIdent(f.Column) + " " + string(op) + " " + expr, args, np
}

// renumber rewrites ? placeholders in expr to the dialect's positional form
// starting at `start`. For Postgres this emits `$N`; for SQLite it emits
// `?N` (numbered) which coexists cleanly with sqlc's SQLite output.
func renumber(expr string, args []any, start int, dialect Dialect) (string, []any, int) {
	if !strings.ContainsRune(expr, '?') {
		return expr, args, start
	}
	prefix := byte('$')
	if dialect == DialectSQLite {
		prefix = '?'
	}
	var sb strings.Builder
	sb.Grow(len(expr) + 4*strings.Count(expr, "?"))
	pos := start
	for _, r := range expr {
		if r == '?' {
			sb.WriteByte(prefix)
			sb.WriteString(strconv.Itoa(pos))
			pos++
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String(), args, pos
}

// quoteIdent double-quotes a Postgres identifier, escaping embedded quotes.
func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
