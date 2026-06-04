package search

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

// Value operators (phase 3). Comparison, range, and wildcard operators are not
// part of the grammar — they ride inside the value string (`<`, `>`, `..`, `*`
// are all valid value characters) and are interpreted here, per the design that
// keeps the grammar tiny and operators a value-coercion concern.

// ValueOp is a value string decomposed into an optional operator and operand(s).
//
//	"42"      -> {Operand: "42"}                 (no operator; caller defaults to =)
//	">=90"    -> {Op: OpGe, Operand: "90"}
//	"10..99"  -> {IsRange: true, Lo: "10", Hi: "99"}
//	"..99"    -> {IsRange: true, Hi: "99"}        (open low → <=)
//	"10.."    -> {IsRange: true, Lo: "10"}        (open high → >=)
type ValueOp struct {
	Op      metaquery.Op // "" when no comparison prefix
	Operand string
	Lo, Hi  string
	IsRange bool
}

// ParseValueOp decomposes a raw value. Range (..) is detected first, then a
// leading comparison operator (two-char before one-char), else a plain operand.
func ParseValueOp(raw string) ValueOp {
	if lo, hi, ok := strings.Cut(raw, ".."); ok {
		return ValueOp{IsRange: true, Lo: lo, Hi: hi}
	}
	for _, p := range []struct {
		s  string
		op metaquery.Op
	}{
		{"<=", metaquery.OpLe}, {">=", metaquery.OpGe}, {"!=", metaquery.OpNe},
		{"<", metaquery.OpLt}, {">", metaquery.OpGt}, {"=", metaquery.OpEq},
	} {
		if strings.HasPrefix(raw, p.s) {
			return ValueOp{Op: p.op, Operand: strings.TrimSpace(raw[len(p.s):])}
		}
	}
	return ValueOp{Operand: raw}
}

// CompareFn builds an Fn for an ordered column (int/float/time/...) that
// understands comparison (<, <=, >, >=, =, !=) and range (a..b, ..b, a..)
// operators. parse coerces an operand string to a bound value; a parse failure
// drops the value (returns nil filter, like the type defaults). A bare value is
// equality.
func CompareFn(parse func(string) (any, error)) Fn {
	return func(c Col, raw string) (*metaquery.Filter, error) {
		vo := ParseValueOp(raw)
		if vo.IsRange {
			lo, loOK := tryParse(parse, vo.Lo)
			hi, hiOK := tryParse(parse, vo.Hi)
			switch {
			case loOK && hiOK:
				return c.Between(lo, hi), nil
			case hiOK:
				return c.Le(hi), nil
			case loOK:
				return c.Ge(lo), nil
			default:
				return nil, nil
			}
		}
		v, err := parse(vo.Operand)
		if err != nil {
			return nil, nil
		}
		op := vo.Op
		if op == "" {
			op = metaquery.OpEq
		}
		return c.cmp(op, v), nil
	}
}

// TimeFn builds a comparison/range Fn for a time column, parsing operands with
// the given layouts (defaults: RFC3339 and date-only "2006-01-02"). Unlocks the
// time kind, which has no type-driven default (date formats are app-specific).
func TimeFn(layouts ...string) Fn {
	if len(layouts) == 0 {
		layouts = []string{time.RFC3339, "2006-01-02"}
	}
	return CompareFn(func(s string) (any, error) {
		for _, l := range layouts {
			if t, err := time.Parse(l, s); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("search: %q not a date", s)
	})
}

// WildcardFn builds a text Fn where * is a glob and a leading = forces exact
// match: "foo*" prefix, "*foo" suffix, "*foo*"/"foo" contains, "=foo" exact.
// Wildcard metacharacters in the value are LIKE-escaped. Postgres uses ILIKE,
// SQLite LIKE.
func WildcardFn() Fn {
	return func(c Col, raw string) (*metaquery.Filter, error) {
		if rest, ok := strings.CutPrefix(raw, "="); ok {
			return c.Eq(rest), nil
		}
		pre := strings.HasPrefix(raw, "*")
		suf := strings.HasSuffix(raw, "*")
		body := raw
		if pre {
			body = body[1:]
		}
		if suf && body != "" {
			body = body[:len(body)-1]
		}
		esc := escapeLike(body)
		var pat string
		switch {
		case pre == suf: // both or neither → contains
			pat = "%" + esc + "%"
		case suf: // foo* → starts with
			pat = esc + "%"
		default: // *foo → ends with
			pat = "%" + esc
		}
		op := "ILIKE"
		if c.dialect == metaquery.DialectSQLite {
			op = "LIKE"
		}
		return &metaquery.Filter{Expr: quoteIdent(c.Column.Name) + " " + op + ` ? ESCAPE '\'`, Args: []any{pat}}, nil
	}
}

func tryParse(parse func(string) (any, error), s string) (any, bool) {
	if s == "" {
		return nil, false
	}
	v, err := parse(s)
	if err != nil {
		return nil, false
	}
	return v, true
}

func coerceInt(s string) (any, error)   { return strconv.ParseInt(s, 10, 64) }
func coerceFloat(s string) (any, error) { return strconv.ParseFloat(s, 64) }
