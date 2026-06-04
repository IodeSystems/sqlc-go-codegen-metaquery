package search

import (
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

func enumQuery() *metaquery.Query {
	return &metaquery.Query{
		Name: "Q",
		SQL:  "SELECT id, name, status FROM t",
		Columns: []metaquery.Column{
			{Name: "id", GoType: "int64", DBType: "integer"},
			{Name: "name", GoType: "string", DBType: "text"},       // plain text
			{Name: "status", GoType: "string", DBType: "order_st"}, // enum-like
		},
	}
}

// whereClause returns the text after "FROM __q WHERE ", or "".
func whereClause(sql string) string {
	_, after, ok := strings.Cut(sql, "FROM __q WHERE ")
	if !ok {
		return ""
	}
	return after
}

func TestApply_EnumExactMatch(t *testing.T) {
	b := metaquery.Wrap(enumQuery())
	if _, err := Apply(b, "status:active", Config{}); err != nil {
		t.Fatal(err)
	}
	sql, args, _ := b.Build()
	if w := whereClause(sql); w != `"status"::text = $1` {
		t.Fatalf("got WHERE %q", w)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Fatalf("args %#v", args)
	}
}

func TestApply_EnumNotGlobal(t *testing.T) {
	// An unqualified term hits plain-text columns (name) but NOT the enum
	// (exact-match types are targeted-only).
	b := metaquery.Wrap(enumQuery())
	if _, err := Apply(b, "active", Config{}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	w := whereClause(sql)
	if strings.Contains(w, "status") {
		t.Fatalf("enum must not be global: %s", w)
	}
	if !strings.Contains(w, `"name" ILIKE`) {
		t.Fatalf("plain text should still be global: %s", w)
	}
}

func TestApply_EnumSQLiteNoCast(t *testing.T) {
	q := enumQuery()
	q.Dialect = metaquery.DialectSQLite
	b := metaquery.Wrap(q)
	if _, err := Apply(b, "status:active", Config{}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if strings.Contains(sql, "::text") {
		t.Fatalf("sqlite should not cast: %s", sql)
	}
	if !strings.Contains(sql, `"status" = ?1`) {
		t.Fatalf("sqlite enum should be plain equality: %s", sql)
	}
}

func TestIsExactText(t *testing.T) {
	plain := []string{"", "text", "varchar", "character varying", "bpchar", "name", "citext", "TEXT"}
	for _, d := range plain {
		if isExactText(d) {
			t.Errorf("%q should be plain text", d)
		}
	}
	for _, d := range []string{"order_status", "uuid", "inet", "mood"} {
		if !isExactText(d) {
			t.Errorf("%q should be exact-match", d)
		}
	}
}
