package metaquery_authors_test

import (
	"strings"
	"testing"

	gen "github.com/iodesystems/sqlc-go-codegen-metaquery/internal/endtoend/testdata/metaquery_authors/gen"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
)

func TestListAuthors_WrapWithPagination(t *testing.T) {
	sql, args, err := metaquery.Wrap(&gen.MetaListAuthors).
		Where("name", "ILIKE", "%foo%").
		OrderBy("id", metaquery.Asc).
		Limit(25).Offset(50).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%foo%" {
		t.Fatalf("args = %v", args)
	}
	checks := []string{
		"WITH __q AS (",
		"SELECT id, name, bio FROM authors",
		"ORDER BY name",
		`"name" ILIKE $1`,
		`ORDER BY "id" ASC`,
		"LIMIT 25 OFFSET 50",
	}
	for _, s := range checks {
		if !strings.Contains(sql, s) {
			t.Errorf("missing %q in:\n%s", s, sql)
		}
	}
}

func TestGetAuthor_PreservesOriginalPlaceholders(t *testing.T) {
	// Typed wrapper: arg type must match at compile time.
	sql, args, err := gen.WrapGetAuthor(7).
		Where("name", "=", "alice").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != int64(7) || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, "WHERE id = $1") {
		t.Fatalf("original $1 must survive inside CTE: %s", sql)
	}
	if !strings.Contains(sql, `"name" = $2`) {
		t.Fatalf("added arg should be $2: %s", sql)
	}
}

func TestWrappers_TypedSignatures(t *testing.T) {
	// Smoke test that all generated wrappers are callable with their expected
	// typed signatures. Compile-time success is the real assertion.
	_ = gen.WrapListAuthors()
	_ = gen.WrapGetAuthor(1)
	_ = gen.WrapDeleteAuthor(2)
	_ = gen.WrapCreateAuthor(gen.CreateAuthorParams{Name: "x"})
}

func TestTypedCols_CompileSafeFilter(t *testing.T) {
	// These calls would fail to compile on any column/op/value mismatch:
	//   gen.ListAuthorsCols.ID.ILike(...)   — IntCol has no ILike
	//   gen.ListAuthorsCols.Name.Eq(42)     — TextCol.Eq wants string
	//   gen.ListAuthorsCols.Typo            — field doesn't exist
	sql, args, err := gen.WrapListAuthors().
		ApplyFilter(gen.ListAuthorsCols.Name.ILike("%ada%")).
		ApplyFilter(gen.ListAuthorsCols.ID.Gt(10)).
		ApplyOrder(gen.ListAuthorsCols.Name.Desc()).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != "%ada%" || args[1] != int64(10) {
		t.Fatalf("args = %v", args)
	}
	for _, want := range []string{`"name" ILIKE $1`, `"id" > $2`, `ORDER BY "name" DESC`} {
		if !strings.Contains(sql, want) {
			t.Errorf("missing %q in: %s", want, sql)
		}
	}
}

func TestValidateFilter_OnGeneratedMeta(t *testing.T) {
	err := metaquery.ValidateFilter(&gen.MetaListAuthors,
		metaquery.Filter{Column: "id", Op: metaquery.OpILike, Value: "x"})
	if err == nil || !strings.Contains(err.Error(), "not valid") {
		t.Fatalf("expected op/type mismatch, got: %v", err)
	}
}

func TestListAuthors_GroupByAggregation(t *testing.T) {
	b := metaquery.Wrap(&gen.MetaListAuthors).
		GroupBy("name").
		Count("total").
		Having("count(*) > ?", 1)
	sql, args, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `SELECT "name", count(*) AS "total" FROM __q`) {
		t.Fatalf("unexpected projection: %s", sql)
	}
	if !strings.Contains(sql, `HAVING count(*) > $1`) {
		t.Fatalf("having missing: %s", sql)
	}

	// OutputColumns should reflect the aggregated projection.
	cols := b.OutputColumns()
	if len(cols) != 2 || cols[0].Name != "name" || cols[1].Name != "total" {
		t.Fatalf("unexpected output columns: %+v", cols)
	}

	// Meta readback: the filter that was applied (Having raw expr) is surfaced.
	meta := b.Meta()
	if len(meta.GroupBy) != 1 || meta.GroupBy[0] != "name" {
		t.Fatalf("meta.GroupBy: %v", meta.GroupBy)
	}
	if len(meta.Having) != 1 || meta.Having[0].Expr == "" {
		t.Fatalf("meta.Having: %+v", meta.Having)
	}
}

type AuthorCount struct {
	Name  string `db:"name"`
	Total int64  `db:"total"`
}

func TestListAuthors_AggregationScanShape(t *testing.T) {
	b := metaquery.Wrap(&gen.MetaListAuthors).
		GroupBy("name").
		Count("total")
	if err := metaquery.Validate[AuthorCount](b); err != nil {
		t.Fatalf("shape should match: %v", err)
	}

	type badShape struct {
		Name string `db:"name"`
	}
	if err := metaquery.Validate[badShape](b); err == nil {
		t.Fatal("expected shape mismatch on missing total")
	}
}

func TestListAuthors_PaginatePopulatesMeta(t *testing.T) {
	b := metaquery.Wrap(&gen.MetaListAuthors).
		ApplyPagination(metaquery.PageRequest{Page: 1, Size: 25, Total: true})
	m := b.Meta()
	if m.Pagination == nil || m.Pagination.Limit != 25 || m.Pagination.Offset != 25 {
		t.Fatalf("pagination: %+v", m.Pagination)
	}
	if !b.WantsTotal() {
		t.Fatal("WantsTotal should be true")
	}
}

func TestDeleteAuthor_NoColumns(t *testing.T) {
	if len(gen.MetaDeleteAuthor.Columns) != 0 {
		t.Fatalf("expected no columns for :exec; got %v", gen.MetaDeleteAuthor.Columns)
	}
	if gen.MetaDeleteAuthor.Cmd != ":exec" {
		t.Fatalf("cmd = %s", gen.MetaDeleteAuthor.Cmd)
	}
}

func TestCreateAuthor_ReturningMetadata(t *testing.T) {
	q := gen.MetaCreateAuthor
	if got := len(q.Columns); got != 3 {
		t.Fatalf("want 3 columns, got %d", got)
	}
	if q.Columns[0].Name != "id" || !q.Columns[0].NotNull {
		t.Fatalf("unexpected id column: %+v", q.Columns[0])
	}
	if q.Table == nil || q.Table.Name != "authors" {
		t.Fatalf("want insert table authors, got %+v", q.Table)
	}
	if got := len(q.Args); got != 2 {
		t.Fatalf("want 2 args, got %d", got)
	}
}
