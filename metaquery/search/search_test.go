package search

import (
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

func sampleQuery() *metaquery.Query {
	return &metaquery.Query{
		Name: "ListUsers",
		SQL:  "SELECT id, name, bio, is_admin FROM users",
		Columns: []metaquery.Column{
			{Name: "id", GoType: "int64", DBType: "int8", NotNull: true},
			{Name: "name", GoType: "string", DBType: "text"},
			{Name: "bio", GoType: "string", DBType: "text"},
			{Name: "is_admin", GoType: "bool", DBType: "bool"},
		},
	}
}

// whereOf builds with the given search applied and returns the WHERE fragment
// (after "FROM __q WHERE ") plus args, or "" when no WHERE was emitted.
func whereOf(t *testing.T, query string, cfg Config) (string, []any) {
	t.Helper()
	b := metaquery.Wrap(sampleQuery())
	if _, err := Apply(b, query, cfg); err != nil {
		t.Fatalf("Apply(%q): %v", query, err)
	}
	sql, args, err := b.Build()
	if err != nil {
		t.Fatalf("Build after %q: %v", query, err)
	}
	const marker = "FROM __q WHERE "
	i := strings.Index(sql, marker)
	if i < 0 {
		return "", args
	}
	return sql[i+len(marker):], args
}

func eqArgs(t *testing.T, got, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args len: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg[%d]: got %#v want %#v", i, got[i], want[i])
		}
	}
}

func TestApply_GlobalText(t *testing.T) {
	where, args := whereOf(t, "john", Config{})
	want := `("name" ILIKE $1 ESCAPE '\' OR "bio" ILIKE $2 ESCAPE '\')`
	if where != want {
		t.Fatalf("where:\n got %s\nwant %s", where, want)
	}
	eqArgs(t, args, []any{"%john%", "%john%"})
}

func TestApply_SpaceIsAnd(t *testing.T) {
	where, _ := whereOf(t, "john smith", Config{})
	if !strings.Contains(where, " AND ") {
		t.Fatalf("space should AND: %s", where)
	}
}

func TestApply_CommaIsOr(t *testing.T) {
	where, _ := whereOf(t, "john, jane", Config{})
	// Two global terms joined by OR.
	if strings.Count(where, " OR ") < 2 { // 1 OR inside each global group + 1 joining terms
		t.Fatalf("comma should OR terms: %s", where)
	}
}

func TestApply_TargetedText(t *testing.T) {
	where, args := whereOf(t, "name:John", Config{})
	want := `"name" ILIKE $1 ESCAPE '\'`
	if where != want {
		t.Fatalf("got %s want %s", where, want)
	}
	eqArgs(t, args, []any{"%John%"})
}

func TestApply_TargetedInt(t *testing.T) {
	where, args := whereOf(t, "id:5", Config{})
	if where != `"id" = $1` {
		t.Fatalf("got %s", where)
	}
	eqArgs(t, args, []any{int64(5)})
}

func TestApply_IntUnparseableDropped(t *testing.T) {
	where, _ := whereOf(t, "id:abc", Config{})
	if where != "" {
		t.Fatalf("unparseable int should drop, got WHERE %s", where)
	}
}

func TestApply_TargetedBool(t *testing.T) {
	where, args := whereOf(t, "is_admin:true", Config{})
	if where != `"is_admin" = $1` {
		t.Fatalf("got %s", where)
	}
	eqArgs(t, args, []any{true})
}

func TestApply_TermNegation(t *testing.T) {
	where, _ := whereOf(t, "!inactive", Config{})
	if !strings.HasPrefix(where, "NOT (") {
		t.Fatalf("term negation should wrap in NOT: %s", where)
	}
}

func TestApply_TargetedNotGlobal(t *testing.T) {
	// is_admin/id are targeted-only: an unqualified term must not touch them.
	where, _ := whereOf(t, "true", Config{})
	if strings.Contains(where, "is_admin") || strings.Contains(where, `"id"`) {
		t.Fatalf("typed cols must be targeted-only: %s", where)
	}
}

func TestApply_UnknownTargetFallsBackToGlobal(t *testing.T) {
	// nosuch:john -> drop target, search "john" globally (dataset behavior).
	where, args := whereOf(t, "nosuch:john", Config{})
	want := `("name" ILIKE $1 ESCAPE '\' OR "bio" ILIKE $2 ESCAPE '\')`
	if where != want {
		t.Fatalf("got %s", where)
	}
	eqArgs(t, args, []any{"%john%", "%john%"})
}

func TestApply_CustomFieldOverride(t *testing.T) {
	cfg := Config{Fields: map[string]Field{
		"id": {Scope: ScopeGlobal, Search: func(c Col, v string) (*metaquery.Filter, error) {
			return c.Gt(v), nil // silly but exercises the closure + Col helper
		}},
	}}
	where, args := whereOf(t, "id:7", cfg)
	if where != `"id" > $1` {
		t.Fatalf("got %s", where)
	}
	eqArgs(t, args, []any{"7"})
}

func TestApply_NamedVirtualSearch(t *testing.T) {
	cfg := Config{Named: map[string]Named{
		"daysago": {Search: func(c Col, v string) (*metaquery.Filter, error) {
			return c.Expr(`"id" > ?`, v), nil
		}},
	}}
	where, args := whereOf(t, "daysAgo:30", cfg) // target match is case-insensitive
	if where != `"id" > $1` {
		t.Fatalf("got %s", where)
	}
	eqArgs(t, args, []any{"30"})
}

func TestApply_Alias(t *testing.T) {
	cfg := Config{Fields: map[string]Field{
		"name": {Aliases: []string{"n"}},
	}}
	where, _ := whereOf(t, "n:bob", cfg)
	if where != `"name" ILIKE $1 ESCAPE '\'` {
		t.Fatalf("alias should resolve to name: %s", where)
	}
}

func TestApply_DisableColumn(t *testing.T) {
	cfg := Config{Fields: map[string]Field{"bio": {Disable: true}}}
	where, _ := whereOf(t, "john", cfg)
	if strings.Contains(where, "bio") {
		t.Fatalf("disabled column must not appear: %s", where)
	}
	if !strings.Contains(where, "name") {
		t.Fatalf("name should still search: %s", where)
	}
}

func TestApply_SQLiteUsesLike(t *testing.T) {
	q := sampleQuery()
	q.Dialect = metaquery.DialectSQLite
	b := metaquery.Wrap(q)
	if _, err := Apply(b, "john", Config{}); err != nil {
		t.Fatal(err)
	}
	sql, _, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sql, "ILIKE") {
		t.Fatalf("sqlite should not use ILIKE: %s", sql)
	}
	if !strings.Contains(sql, "LIKE ?1") {
		t.Fatalf("sqlite should use numbered LIKE placeholders: %s", sql)
	}
}

func TestApply_Empty(t *testing.T) {
	where, _ := whereOf(t, "   ", Config{})
	if where != "" {
		t.Fatalf("empty search should add no WHERE: %s", where)
	}
}

func TestApply_LikeWildcardsEscaped(t *testing.T) {
	_, args := whereOf(t, `name:50%`, Config{})
	if len(args) != 1 || args[0] != `%50\%%` {
		t.Fatalf("user %% should be escaped: %#v", args)
	}
}

// --- Parser-level conformance ---

func TestParse_QuotedPreservesSpaces(t *testing.T) {
	p, err := Parse(`name:"John Doe"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Terms) != 1 || len(p.Terms[0].Values) != 1 || p.Terms[0].Values[0].Value != "John Doe" {
		t.Fatalf("quoted value not preserved: %+v", p.Terms)
	}
	if p.Terms[0].Target != "name" {
		t.Fatalf("target: %q", p.Terms[0].Target)
	}
}

func TestParse_GroupedValuesOr(t *testing.T) {
	p, err := Parse("status:(a, b)")
	if err != nil {
		t.Fatal(err)
	}
	vs := p.Terms[0].Values
	if len(vs) != 2 || vs[0].Value != "a" || vs[1].Value != "b" || vs[1].Conj != Or {
		t.Fatalf("grouped OR values: %+v", vs)
	}
}

func TestParse_GroupedValuesAnd(t *testing.T) {
	p, err := Parse("status:(a b)")
	if err != nil {
		t.Fatal(err)
	}
	vs := p.Terms[0].Values
	if len(vs) != 2 || vs[1].Conj != And {
		t.Fatalf("grouped AND values: %+v", vs)
	}
}

func TestParse_AutoEscapeRecovery(t *testing.T) {
	// A lone special char that can't parse gets backslash-escaped and retried;
	// the result must be stable (idempotent) and re-parse without further change.
	p, err := Parse("(")
	if err != nil {
		t.Fatalf("recovery should not error: %v", err)
	}
	p2, err := Parse(p.Search)
	if err != nil {
		t.Fatal(err)
	}
	if p2.Search != p.Search {
		t.Fatalf("rendered not idempotent: %q -> %q", p.Search, p2.Search)
	}
}

func TestParse_BackslashEscape(t *testing.T) {
	p, err := Parse(`a\:b`)
	if err != nil {
		t.Fatal(err)
	}
	// escaped colon -> literal value "a:b", no target.
	if len(p.Terms) != 1 || p.Terms[0].Target != "" || p.Terms[0].Values[0].Value != "a:b" {
		t.Fatalf("backslash escape: %+v", p.Terms)
	}
}
