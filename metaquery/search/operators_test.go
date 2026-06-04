package search

import (
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

func TestParseValueOp(t *testing.T) {
	cases := []struct {
		in   string
		want ValueOp
	}{
		{"42", ValueOp{Operand: "42"}},
		{">5", ValueOp{Op: metaquery.OpGt, Operand: "5"}},
		{">=90", ValueOp{Op: metaquery.OpGe, Operand: "90"}},
		{"<=9", ValueOp{Op: metaquery.OpLe, Operand: "9"}},
		{"<7", ValueOp{Op: metaquery.OpLt, Operand: "7"}},
		{"=3", ValueOp{Op: metaquery.OpEq, Operand: "3"}},
		{"10..99", ValueOp{IsRange: true, Lo: "10", Hi: "99"}},
		{"..99", ValueOp{IsRange: true, Hi: "99"}},
		{"10..", ValueOp{IsRange: true, Lo: "10"}},
		{"1.5", ValueOp{Operand: "1.5"}}, // single dot is not a range
	}
	for _, c := range cases {
		if got := ParseValueOp(c.in); got != c.want {
			t.Errorf("ParseValueOp(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

// Numeric defaults are operator-aware (id is int64 in sampleQuery).
func TestApply_IntComparison(t *testing.T) {
	where, args := whereOf(t, "id:>5", Config{})
	if where != `"id" > $1` {
		t.Fatalf("got %s", where)
	}
	if len(args) != 1 || args[0] != int64(5) {
		t.Fatalf("args %#v", args)
	}
}

func TestApply_IntRange(t *testing.T) {
	where, args := whereOf(t, "id:5..9", Config{})
	if where != `"id" BETWEEN $1 AND $2` {
		t.Fatalf("got %s", where)
	}
	if len(args) != 2 || args[0] != int64(5) || args[1] != int64(9) {
		t.Fatalf("args %#v", args)
	}
}

func TestApply_IntOpenRange(t *testing.T) {
	if where, args := whereOf(t, "id:..9", Config{}); where != `"id" <= $1` || args[0] != int64(9) {
		t.Fatalf("..9 got %s %v", where, args)
	}
	if where, args := whereOf(t, "id:5..", Config{}); where != `"id" >= $1` || args[0] != int64(5) {
		t.Fatalf("5.. got %s %v", where, args)
	}
}

func TestApply_IntEqStillWorks(t *testing.T) {
	// Bare value keeps equality (regression: operator-aware default must not
	// change the common case).
	if where, args := whereOf(t, "id:7", Config{}); where != `"id" = $1` || args[0] != int64(7) {
		t.Fatalf("got %s %v", where, args)
	}
}

func TestApply_IntComparisonUnparseableDropped(t *testing.T) {
	if where, _ := whereOf(t, "id:>abc", Config{}); where != "" {
		t.Fatalf("unparseable operand should drop: %s", where)
	}
}

// WildcardFn / TimeFn exercised directly (sampleQuery has no time column).
func col(name string, d metaquery.Dialect) Col {
	return Col{Column: metaquery.Column{Name: name}, dialect: d}
}

func TestWildcardFn(t *testing.T) {
	fn := WildcardFn()
	cases := []struct {
		in, expr string
		arg      any
	}{
		{"foo", `"name" ILIKE ? ESCAPE '\'`, "%foo%"},
		{"foo*", `"name" ILIKE ? ESCAPE '\'`, "foo%"},
		{"*foo", `"name" ILIKE ? ESCAPE '\'`, "%foo"},
		{"*foo*", `"name" ILIKE ? ESCAPE '\'`, "%foo%"},
		{"=foo", "", "foo"}, // exact → structured Eq, not expr
	}
	for _, c := range cases {
		f, err := fn(col("name", metaquery.DialectPostgres), c.in)
		if err != nil || f == nil {
			t.Fatalf("WildcardFn(%q): %v", c.in, err)
		}
		if c.expr == "" { // exact match path
			if f.Op != metaquery.OpEq || f.Value != c.arg {
				t.Errorf("=foo: got op=%q val=%v", f.Op, f.Value)
			}
			continue
		}
		if f.Expr != c.expr || len(f.Args) != 1 || f.Args[0] != c.arg {
			t.Errorf("WildcardFn(%q) = {%q,%v}, want {%q,%v}", c.in, f.Expr, f.Args, c.expr, c.arg)
		}
	}
}

func TestWildcardFn_SQLiteUsesLike(t *testing.T) {
	f, _ := WildcardFn()(col("name", metaquery.DialectSQLite), "foo*")
	if f.Expr != `"name" LIKE ? ESCAPE '\'` {
		t.Fatalf("sqlite wildcard: %s", f.Expr)
	}
}

func TestTimeFn(t *testing.T) {
	fn := TimeFn() // default layouts incl. "2006-01-02"
	// range
	f, err := fn(col("created", metaquery.DialectPostgres), "2024-01-01..2024-12-31")
	if err != nil || f == nil {
		t.Fatalf("range: %v", err)
	}
	if f.Expr != `"created" BETWEEN ? AND ?` || len(f.Args) != 2 {
		t.Fatalf("range expr: %s args=%v", f.Expr, f.Args)
	}
	// comparison
	f, _ = fn(col("created", metaquery.DialectPostgres), ">=2024-06-01")
	if f.Column != "created" || f.Op != metaquery.OpGe {
		t.Fatalf("cmp: %+v", f)
	}
	// unparseable date → dropped
	if f, _ := fn(col("created", metaquery.DialectPostgres), "notadate"); f != nil {
		t.Fatalf("bad date should drop: %+v", f)
	}
}
