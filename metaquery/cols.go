package metaquery

import "time"

// Column-reference kinds. Each kind carries a column name and exposes op
// methods appropriate to its Go type. The plugin emits a
// `<Method>Cols` struct per query whose fields are these kinds, so
// callers get compile-time-safe column names and value types:
//
//	b.ApplyFilter(db.ListAuthorsCols.Name.ILike("%foo%"))
//	b.ApplyOrder(db.ListAuthorsCols.CreatedAt.Desc())
//
// Constructors (NewTextCol, NewIntCol, ...) are used by generated code so
// struct literals stay one-liners.

// baseCol carries the column name and provides kind-agnostic methods
// (null checks, ordering). Embedded by every typed kind.
type baseCol struct {
	name string
}

// Col returns the column name this reference points at. Useful for
// forwarding to GroupBy/Select which take plain strings:
//
//	b.GroupBy(db.ListAuthorsCols.Name.Col())
func (c baseCol) Col() string { return c.name }

// IsNull / IsNotNull emit a value-less WHERE/HAVING fragment via the raw-expr
// path (OpIsNull/OpIsNotNull on Filter renders the same, but IS NULL rarely
// reads as well through the structured path).
func (c baseCol) IsNull() Filter {
	return Filter{Column: c.name, Op: OpIsNull}
}

func (c baseCol) IsNotNull() Filter {
	return Filter{Column: c.name, Op: OpIsNotNull}
}

// Asc / Desc produce OrderBy entries for the builder's ApplyOrder method.
func (c baseCol) Asc() OrderBy  { return OrderBy{Column: c.name, Dir: Asc} }
func (c baseCol) Desc() OrderBy { return OrderBy{Column: c.name, Dir: Desc} }

// TextCol wraps string / pgtype.Text / pgtype.Varchar columns.
type TextCol struct{ baseCol }

func NewTextCol(name string) TextCol { return TextCol{baseCol{name}} }

func (c TextCol) Eq(v string) Filter    { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c TextCol) Ne(v string) Filter    { return Filter{Column: c.name, Op: OpNe, Value: v} }
func (c TextCol) Like(v string) Filter  { return Filter{Column: c.name, Op: OpLike, Value: v} }
func (c TextCol) ILike(v string) Filter { return Filter{Column: c.name, Op: OpILike, Value: v} }

func (c TextCol) In(vs ...string) Filter {
	if len(vs) == 0 {
		return Filter{Expr: "FALSE"}
	}
	return Filter{Expr: quoteIdent(c.name) + " = ANY(?)", Args: []any{vs}}
}

// IntCol wraps int/int16/int32/int64 and pgtype.Int2/4/8. Values are taken
// as int64; Go's untyped integer constants make callsites ergonomic
// (c.Id.Eq(42) compiles because 42 is untyped).
type IntCol struct{ baseCol }

func NewIntCol(name string) IntCol { return IntCol{baseCol{name}} }

func (c IntCol) Eq(v int64) Filter { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c IntCol) Ne(v int64) Filter { return Filter{Column: c.name, Op: OpNe, Value: v} }
func (c IntCol) Gt(v int64) Filter { return Filter{Column: c.name, Op: OpGt, Value: v} }
func (c IntCol) Ge(v int64) Filter { return Filter{Column: c.name, Op: OpGe, Value: v} }
func (c IntCol) Lt(v int64) Filter { return Filter{Column: c.name, Op: OpLt, Value: v} }
func (c IntCol) Le(v int64) Filter { return Filter{Column: c.name, Op: OpLe, Value: v} }

func (c IntCol) In(vs ...int64) Filter {
	if len(vs) == 0 {
		return Filter{Expr: "FALSE"}
	}
	return Filter{Expr: quoteIdent(c.name) + " = ANY(?)", Args: []any{vs}}
}

func (c IntCol) Between(lo, hi int64) Filter {
	return Filter{Expr: quoteIdent(c.name) + " BETWEEN ? AND ?", Args: []any{lo, hi}}
}

// FloatCol wraps float32/float64 and pgtype.Float4/8.
type FloatCol struct{ baseCol }

func NewFloatCol(name string) FloatCol { return FloatCol{baseCol{name}} }

func (c FloatCol) Eq(v float64) Filter { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c FloatCol) Ne(v float64) Filter { return Filter{Column: c.name, Op: OpNe, Value: v} }
func (c FloatCol) Gt(v float64) Filter { return Filter{Column: c.name, Op: OpGt, Value: v} }
func (c FloatCol) Ge(v float64) Filter { return Filter{Column: c.name, Op: OpGe, Value: v} }
func (c FloatCol) Lt(v float64) Filter { return Filter{Column: c.name, Op: OpLt, Value: v} }
func (c FloatCol) Le(v float64) Filter { return Filter{Column: c.name, Op: OpLe, Value: v} }

func (c FloatCol) Between(lo, hi float64) Filter {
	return Filter{Expr: quoteIdent(c.name) + " BETWEEN ? AND ?", Args: []any{lo, hi}}
}

// BoolCol wraps bool / pgtype.Bool. IsTrue / IsFalse read nicer than Eq(true).
type BoolCol struct{ baseCol }

func NewBoolCol(name string) BoolCol { return BoolCol{baseCol{name}} }

func (c BoolCol) Eq(v bool) Filter { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c BoolCol) IsTrue() Filter   { return Filter{Column: c.name, Op: OpEq, Value: true} }
func (c BoolCol) IsFalse() Filter  { return Filter{Column: c.name, Op: OpEq, Value: false} }

// TimeCol wraps time.Time and pgtype.Timestamp/Timestamptz/Date/Time.
type TimeCol struct{ baseCol }

func NewTimeCol(name string) TimeCol { return TimeCol{baseCol{name}} }

func (c TimeCol) Eq(v time.Time) Filter     { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c TimeCol) Before(v time.Time) Filter { return Filter{Column: c.name, Op: OpLt, Value: v} }
func (c TimeCol) After(v time.Time) Filter  { return Filter{Column: c.name, Op: OpGt, Value: v} }

func (c TimeCol) Between(lo, hi time.Time) Filter {
	return Filter{Expr: quoteIdent(c.name) + " BETWEEN ? AND ?", Args: []any{lo, hi}}
}

// BytesCol wraps []byte columns.
type BytesCol struct{ baseCol }

func NewBytesCol(name string) BytesCol { return BytesCol{baseCol{name}} }

func (c BytesCol) Eq(v []byte) Filter { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c BytesCol) Ne(v []byte) Filter { return Filter{Column: c.name, Op: OpNe, Value: v} }

// AnyCol is the fallback for column kinds we don't have dedicated typing
// for (arrays, enums, pgtype.UUID, pgtype.Numeric, domain types, ...). Ops
// take `any` — use sparingly.
type AnyCol struct{ baseCol }

func NewAnyCol(name string) AnyCol { return AnyCol{baseCol{name}} }

func (c AnyCol) Eq(v any) Filter { return Filter{Column: c.name, Op: OpEq, Value: v} }
func (c AnyCol) Ne(v any) Filter { return Filter{Column: c.name, Op: OpNe, Value: v} }
func (c AnyCol) Gt(v any) Filter { return Filter{Column: c.name, Op: OpGt, Value: v} }
func (c AnyCol) Ge(v any) Filter { return Filter{Column: c.name, Op: OpGe, Value: v} }
func (c AnyCol) Lt(v any) Filter { return Filter{Column: c.name, Op: OpLt, Value: v} }
func (c AnyCol) Le(v any) Filter { return Filter{Column: c.name, Op: OpLe, Value: v} }
