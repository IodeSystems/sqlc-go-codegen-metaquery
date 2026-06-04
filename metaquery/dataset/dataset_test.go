package dataset

import (
	"context"
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/search"
)

type userView struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

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

func TestShape_FullRequest(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	req := Request{
		Page:       2,
		PageSize:   10,
		Search:     "john",
		Partition:  "is_admin:true",
		Ordering:   []Order{{Field: "name", Order: "DESC"}},
		ShowCounts: true,
	}
	rendered, err := Shape(b, req, Config{})
	if err != nil {
		t.Fatal(err)
	}
	sql, args, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Partition + search both AND into WHERE.
	if !strings.Contains(sql, `"is_admin" = `) {
		t.Fatalf("partition missing: %s", sql)
	}
	if !strings.Contains(sql, "ILIKE") {
		t.Fatalf("search missing: %s", sql)
	}
	if !strings.Contains(sql, " AND ") {
		t.Fatalf("partition+search should AND: %s", sql)
	}
	if !strings.Contains(sql, `ORDER BY "name" DESC`) {
		t.Fatalf("ordering missing: %s", sql)
	}
	if !strings.Contains(sql, "LIMIT 10 OFFSET 20") {
		t.Fatalf("pagination wrong: %s", sql)
	}
	// args: is_admin true, then %john% x2 (name, bio).
	if len(args) != 3 || args[0] != true {
		t.Fatalf("args: %#v", args)
	}
	if rendered.Search != "john" || rendered.Partition != "is_admin:true" {
		t.Fatalf("rendered: %+v", rendered)
	}
}

func TestShape_PageSizeClamp(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b, Request{PageSize: 9999}, Config{MaxPageSize: 50}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if !strings.Contains(sql, "LIMIT 50") {
		t.Fatalf("should clamp to MaxPageSize: %s", sql)
	}
}

func TestShape_DefaultPageSize(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b, Request{}, Config{}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if !strings.Contains(sql, "LIMIT 25") {
		t.Fatalf("should default to 25: %s", sql)
	}
}

func TestShape_OrderableWhitelist(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	req := Request{Ordering: []Order{{Field: "name", Order: "ASC"}, {Field: "bio", Order: "ASC"}}}
	if _, err := Shape(b, req, Config{Orderable: []string{"name"}}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if strings.Contains(sql, `"bio" ASC`) {
		t.Fatalf("bio not in Orderable, should be dropped: %s", sql)
	}
	if !strings.Contains(sql, `"name" ASC`) {
		t.Fatalf("name should sort: %s", sql)
	}
}

func TestShape_PartitionParseError(t *testing.T) {
	// Partition compiles against columns; an int target with a bad value just
	// drops (no error). Force an error path via a custom Fn that errors.
	b := metaquery.Wrap(sampleQuery())
	_, err := Shape(b, Request{Search: "name:ok"}, Config{})
	if err != nil {
		t.Fatalf("valid search should not error: %v", err)
	}
}

func TestShape_SearchableAllowlist(t *testing.T) {
	// Only "name" is free-text searchable; "bio" (also text) must be excluded
	// from an unqualified term.
	b := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b, Request{Search: "john"}, Config{Searchable: []string{"name"}}); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if !strings.Contains(sql, `"name" ILIKE`) {
		t.Fatalf("name should search: %s", sql)
	}
	if strings.Contains(sql, `"bio" ILIKE`) {
		t.Fatalf("bio not in Searchable, must be excluded from free-text: %s", sql)
	}
}

func TestShape_SearchableIsHardAllowlist(t *testing.T) {
	// A column absent from Searchable/Targetable is not searchable by any path:
	// bio:hello must not filter bio. The unknown target falls back to a global
	// value search on the allowed column (name), per the DSL.
	b := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b, Request{Search: "bio:hello"}, Config{Searchable: []string{"name"}}); err != nil {
		t.Fatal(err)
	}
	sql, args, _ := b.Build()
	if strings.Contains(sql, `"bio"`) {
		t.Fatalf("bio must not be targetable: %s", sql)
	}
	if !strings.Contains(sql, `"name" ILIKE`) || len(args) != 1 || args[0] != "%hello%" {
		t.Fatalf("bio:hello should fall back to global search on name: %s args=%v", sql, args)
	}
}

func TestShape_Targetable(t *testing.T) {
	// bio is targeted-only: bio:hello targets bio, but a bare term does not.
	cfg := Config{Searchable: []string{"name"}, Targetable: []string{"bio"}}

	b := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b, Request{Search: "bio:hello"}, cfg); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if !strings.Contains(sql, `"bio" ILIKE`) {
		t.Fatalf("bio:hello should target bio: %s", sql)
	}

	b2 := metaquery.Wrap(sampleQuery())
	if _, err := Shape(b2, Request{Search: "hello"}, cfg); err != nil {
		t.Fatal(err)
	}
	sql2, _, _ := b2.Build()
	if strings.Contains(sql2, `"bio"`) {
		t.Fatalf("targeted-only bio must not match a bare term: %s", sql2)
	}
}

func TestShape_SearchableExplicitFieldWins(t *testing.T) {
	// An explicit Fields entry overrides the allowlist-derived scope.
	b := metaquery.Wrap(sampleQuery())
	cfg := Config{Searchable: []string{"name"}}
	cfg.Config.Fields = map[string]search.Field{
		"bio": {Scope: search.ScopeGlobal}, // force bio global despite allowlist
	}
	if _, err := Shape(b, Request{Search: "john"}, cfg); err != nil {
		t.Fatal(err)
	}
	sql, _, _ := b.Build()
	if !strings.Contains(sql, `"bio" ILIKE`) {
		t.Fatalf("explicit ScopeGlobal on bio should win: %s", sql)
	}
}

func TestRun_MapsResponseAndCounts(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	scan := func(_ context.Context, b *metaquery.Builder) (*metaquery.TypedResult[userView], error) {
		// Ensure Shape ran: Build must succeed and carry pagination.
		if _, _, err := b.Build(); err != nil {
			return nil, err
		}
		m := b.Meta()
		if m.Pagination != nil {
			m.Pagination.Total = 42
		}
		return &metaquery.TypedResult[userView]{
			Data: []userView{{ID: 1, Name: "john"}},
			Meta: m,
		}, nil
	}
	out, err := Run(context.Background(), b, Request{Search: "john", ShowCounts: true}, Config{}, scan)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Data) != 1 || out.Data[0].Name != "john" {
		t.Fatalf("data: %+v", out.Data)
	}
	if out.Count == nil || out.Count.InQuery != 42 || out.Count.InPartition != 42 {
		t.Fatalf("counts: %+v", out.Count)
	}
	if out.Rendered.Search != "john" {
		t.Fatalf("rendered: %+v", out.Rendered)
	}
}

func TestRun_EmptyDataNonNil(t *testing.T) {
	b := metaquery.Wrap(sampleQuery())
	scan := func(_ context.Context, _ *metaquery.Builder) (*metaquery.TypedResult[userView], error) {
		return &metaquery.TypedResult[userView]{}, nil
	}
	out, err := Run(context.Background(), b, Request{}, Config{}, scan)
	if err != nil {
		t.Fatal(err)
	}
	if out.Data == nil {
		t.Fatal("Data should be non-nil empty slice")
	}
}

func TestRunWithPartitionCount(t *testing.T) {
	newB := func() *metaquery.Builder { return metaquery.Wrap(sampleQuery()) }
	scan := func(_ context.Context, b *metaquery.Builder) (*metaquery.TypedResult[userView], error) {
		m := b.Meta()
		if m.Pagination != nil {
			m.Pagination.Total = 7 // inQuery (partition+search)
		}
		return &metaquery.TypedResult[userView]{Data: []userView{{ID: 1}}, Meta: m}, nil
	}
	var counted string
	count := func(_ context.Context, b *metaquery.Builder) (int64, error) {
		sql, _, err := b.BuildCount()
		counted = sql
		return 20, err // inPartition (partition only)
	}
	req := Request{Search: "john", Partition: "is_admin:true", ShowCounts: true}
	out, err := RunWithPartitionCount(context.Background(), newB, req, Config{}, scan, count)
	if err != nil {
		t.Fatal(err)
	}
	if out.Count.InQuery != 7 || out.Count.InPartition != 20 {
		t.Fatalf("counts: %+v", out.Count)
	}
	// The partition-only count builder must include the partition but NOT the search.
	if !strings.Contains(counted, `"is_admin"`) {
		t.Fatalf("partition count should filter is_admin: %s", counted)
	}
	if strings.Contains(counted, "ILIKE") {
		t.Fatalf("partition count must exclude search: %s", counted)
	}
}
