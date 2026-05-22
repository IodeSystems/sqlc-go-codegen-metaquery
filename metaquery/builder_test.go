package metaquery

import (
	"strings"
	"testing"
)

var listAuthors = &Query{
	Name: "ListAuthors",
	Cmd:  ":many",
	SQL:  "SELECT id, name, bio FROM authors ORDER BY name",
	Columns: []Column{
		{Name: "id", OriginalName: "id", GoType: "int64", DBType: "int8", NotNull: true},
		{Name: "name", OriginalName: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", OriginalName: "bio", GoType: "pgtype.Text", DBType: "text"},
	},
}

var getAuthor = &Query{
	Name: "GetAuthor",
	Cmd:  ":one",
	SQL:  "SELECT id, name, bio FROM authors WHERE id = $1 LIMIT 1",
	Columns: []Column{
		{Name: "id", OriginalName: "id", GoType: "int64", DBType: "int8", NotNull: true},
		{Name: "name", OriginalName: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", OriginalName: "bio", GoType: "pgtype.Text", DBType: "text"},
	},
	Args: []Arg{{Position: 1, GoType: "int64", DBType: "int8", NotNull: true}},
}

func TestBuilder_BareWrap(t *testing.T) {
	sql, args, err := Wrap(listAuthors).Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	want := "WITH __q AS (\nSELECT id, name, bio FROM authors ORDER BY name\n)\nSELECT * FROM __q"
	if sql != want {
		t.Fatalf("sql mismatch:\n got: %q\nwant: %q", sql, want)
	}
}

func TestBuilder_WhereLimitOffset(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		Where("name", "ILIKE", "%foo%").
		OrderBy("name", Asc).
		Limit(50).Offset(100).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%foo%" {
		t.Fatalf("args = %v", args)
	}
	wantSub := `"name" ILIKE $1`
	if !strings.Contains(sql, wantSub) {
		t.Fatalf("missing %q in %q", wantSub, sql)
	}
	if !strings.Contains(sql, "LIMIT 50 OFFSET 100") {
		t.Fatalf("missing LIMIT/OFFSET in %q", sql)
	}
	if !strings.Contains(sql, `ORDER BY "name" ASC`) {
		t.Fatalf("missing ORDER BY in %q", sql)
	}
}

func TestBuilder_RenumbersOverBaseArgs(t *testing.T) {
	sql, args, err := Wrap(getAuthor, int64(42)).
		Where("name", "=", "alice").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != int64(42) || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"name" = $2`) {
		t.Fatalf("expected $2 for added arg, got: %q", sql)
	}
	if !strings.Contains(sql, "WHERE id = $1") {
		t.Fatalf("original $1 should be preserved inside CTE: %q", sql)
	}
}

func TestBuilder_WhereRejectsUnknownColumn(t *testing.T) {
	_, _, err := Wrap(listAuthors).Where("nope", "=", 1).Build()
	if err == nil {
		t.Fatal("expected whitelist error")
	}
	if !strings.Contains(err.Error(), `unknown column "nope"`) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBuilder_GroupByAgg(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		Select("name"). // ignored in favor of groupBy/aggs
		GroupBy("name").
		Agg("total", "count(*)", "int64").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `SELECT "name", count(*) AS "total" FROM __q`) {
		t.Fatalf("unexpected sql: %q", sql)
	}
	if !strings.Contains(sql, `GROUP BY "name"`) {
		t.Fatalf("missing group by: %q", sql)
	}
}

func TestBuilder_WhereExprPassthrough(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		WhereExpr("name ILIKE ? OR bio ILIKE ?", "%a%", "%b%").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, "name ILIKE $1 OR bio ILIKE $2") {
		t.Fatalf("unexpected sql: %q", sql)
	}
}

func TestBuilder_SelectProjection(t *testing.T) {
	sql, _, err := Wrap(listAuthors).Select("id", "name").Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `SELECT "id", "name" FROM __q`) {
		t.Fatalf("unexpected sql: %q", sql)
	}
}

func TestBuilder_HavingWithAgg(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		GroupBy("name").
		Agg("total", "count(*)", "int64").
		Having("count(*) > ?", 5).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != 5 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, "HAVING count(*) > $1") {
		t.Fatalf("unexpected sql: %q", sql)
	}
}

func TestBuilder_TypedAggHelpers(t *testing.T) {
	b := Wrap(listAuthors).GroupBy("name").Count("total").Sum("max_id", "id")
	sql, _, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `count(*) AS "total"`) {
		t.Fatalf("count missing: %q", sql)
	}
	if !strings.Contains(sql, `sum("id") AS "max_id"`) {
		t.Fatalf("sum missing: %q", sql)
	}
	cols := b.OutputColumns()
	if len(cols) != 3 {
		t.Fatalf("want 3 output cols, got %d", len(cols))
	}
	if cols[1].Name != "total" || cols[1].GoType != "int64" {
		t.Fatalf("unexpected total column: %+v", cols[1])
	}
	if cols[2].Name != "max_id" || cols[2].GoType != "int64" {
		t.Fatalf("unexpected max_id column: %+v", cols[2])
	}
}

func TestBuilder_OutputColumnsPassthrough(t *testing.T) {
	cols := Wrap(listAuthors).OutputColumns()
	if len(cols) != 3 {
		t.Fatalf("want 3 cols, got %d", len(cols))
	}
	if cols[0].Name != "id" || cols[0].GoType != "int64" {
		t.Fatalf("unexpected id col: %+v", cols[0])
	}
}

func TestBuilder_BuildCountSimple(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		Where("name", "=", "x").
		Limit(10).Offset(5).
		OrderBy("id", Desc).
		BuildCount()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "SELECT count(*)") {
		t.Fatalf("missing count in: %q", sql)
	}
	// Outer query (after CTE close) must not contain pagination/order.
	_, outer, ok := strings.Cut(sql, ")\n")
	if !ok {
		t.Fatalf("no CTE boundary in: %q", sql)
	}
	if strings.Contains(outer, "LIMIT") || strings.Contains(outer, "OFFSET") || strings.Contains(outer, "ORDER BY") {
		t.Fatalf("count query must strip pagination/order: %q", outer)
	}
	if len(args) != 1 {
		t.Fatalf("args = %v", args)
	}
}

func TestBuilder_BuildCountWithGroupBy(t *testing.T) {
	sql, _, err := Wrap(listAuthors).
		GroupBy("name").
		Count("total").
		BuildCount()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "SELECT count(*) FROM (SELECT") {
		t.Fatalf("group-by count should wrap projection: %q", sql)
	}
	if !strings.Contains(sql, ") __c") {
		t.Fatalf("missing count alias: %q", sql)
	}
}

func TestBuilder_PaginateWithTotal(t *testing.T) {
	b := Wrap(listAuthors).ApplyPagination(PageRequest{Page: 2, Size: 20, Total: true})
	if !b.WantsTotal() {
		t.Fatal("WantsTotal should be true")
	}
	meta := b.Meta()
	if meta.Pagination == nil || meta.Pagination.Limit != 20 || meta.Pagination.Offset != 40 {
		t.Fatalf("bad pagination: %+v", meta.Pagination)
	}
}

func TestBuilder_FilterReadback(t *testing.T) {
	b := Wrap(listAuthors).
		ApplyFilter(Filter{Column: "name", Op: "ILIKE", Value: "%a%"}).
		ApplyOrder(OrderBy{Column: "id", Dir: Desc})
	_, _, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	meta := b.Meta()
	if len(meta.Where) != 1 || meta.Where[0].Column != "name" || meta.Where[0].Op != "ILIKE" {
		t.Fatalf("where readback: %+v", meta.Where)
	}
	if len(meta.OrderBy) != 1 || meta.OrderBy[0].Column != "id" || meta.OrderBy[0].Dir != Desc {
		t.Fatalf("order readback: %+v", meta.OrderBy)
	}
}

func TestValidate_Match(t *testing.T) {
	type row struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Bio  string `db:"bio"`
	}
	b := Wrap(listAuthors)
	if err := Validate[row](b); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidate_ExtraField(t *testing.T) {
	type row struct {
		ID    int64  `db:"id"`
		Name  string `db:"name"`
		Extra string `db:"extra"`
	}
	b := Wrap(listAuthors)
	err := Validate[row](b)
	if err == nil || !strings.Contains(err.Error(), `column "extra"`) {
		t.Fatalf("expected extra-field error, got: %v", err)
	}
}

func TestValidate_MissingField(t *testing.T) {
	type row struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
	}
	b := Wrap(listAuthors)
	err := Validate[row](b)
	if err == nil || !strings.Contains(err.Error(), `"bio"`) {
		t.Fatalf("expected missing-column error, got: %v", err)
	}
}

func TestValidate_AggShape(t *testing.T) {
	type row struct {
		Name  string `db:"name"`
		Total int64  `db:"total"`
	}
	b := Wrap(listAuthors).GroupBy("name").Count("total")
	if err := Validate[row](b); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ---- Op enum + typed cols + ValidateFilter ----

var nameCol = NewTextCol("name")
var idCol = NewIntCol("id")

func TestOpConstants_UsedInWhere(t *testing.T) {
	sql, _, err := Wrap(listAuthors).Where("name", OpILike, "%a%").Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `"name" ILIKE $1`) {
		t.Fatalf("unexpected: %q", sql)
	}
}

func TestTypedCol_TextILike(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		ApplyFilter(nameCol.ILike("%foo%")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%foo%" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"name" ILIKE $1`) {
		t.Fatalf("unexpected: %q", sql)
	}
}

func TestTypedCol_IntBetween(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		ApplyFilter(idCol.Between(1, 100)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != int64(1) || args[1] != int64(100) {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"id" BETWEEN $1 AND $2`) {
		t.Fatalf("unexpected: %q", sql)
	}
}

func TestTypedCol_AscDesc(t *testing.T) {
	sql, _, err := Wrap(listAuthors).
		ApplyOrder(nameCol.Desc()).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, `ORDER BY "name" DESC`) {
		t.Fatalf("unexpected: %q", sql)
	}
}

func TestTypedCol_IsNullEmitsNoPlaceholder(t *testing.T) {
	sql, args, err := Wrap(listAuthors).
		ApplyFilter(NewTextCol("bio").IsNull()).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"bio" IS NULL`) {
		t.Fatalf("unexpected: %q", sql)
	}
}

func TestValidateFilter_OpTypeMismatch(t *testing.T) {
	err := ValidateFilter(listAuthors, Filter{Column: "id", Op: OpILike, Value: "x"})
	if err == nil || !strings.Contains(err.Error(), "not valid") {
		t.Fatalf("expected op/type error, got: %v", err)
	}
}

func TestValidateFilter_ValueTypeMismatch(t *testing.T) {
	err := ValidateFilter(listAuthors, Filter{Column: "id", Op: OpEq, Value: "not an int"})
	if err == nil || !strings.Contains(err.Error(), "want int") {
		t.Fatalf("expected value type error, got: %v", err)
	}
}

func TestValidateFilter_JSONNumberAcceptedForInt(t *testing.T) {
	// JSON numbers unmarshal as float64 — should pass for int columns if integral.
	if err := ValidateFilter(listAuthors, Filter{Column: "id", Op: OpEq, Value: float64(42)}); err != nil {
		t.Fatalf("integral float64 should be accepted: %v", err)
	}
	if err := ValidateFilter(listAuthors, Filter{Column: "id", Op: OpEq, Value: float64(42.5)}); err == nil {
		t.Fatal("non-integral float64 should be rejected")
	}
}

func TestValidateFilter_UnknownColumn(t *testing.T) {
	err := ValidateFilter(listAuthors, Filter{Column: "nope", Op: OpEq, Value: 1})
	if err == nil || !strings.Contains(err.Error(), "unknown column") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateFilter_RawExprPassthrough(t *testing.T) {
	// Raw-expr filters bypass validation — caller owns safety.
	if err := ValidateFilter(listAuthors, Filter{Expr: "anything goes"}); err != nil {
		t.Fatalf("raw expr should pass: %v", err)
	}
}

// ---- SQLite dialect ----

var listAuthorsSQLite = &Query{
	Name:    "ListAuthors",
	Cmd:     ":many",
	Dialect: DialectSQLite,
	SQL:     "SELECT id, name, bio FROM authors ORDER BY name",
	Columns: []Column{
		{Name: "id", OriginalName: "id", GoType: "int64", DBType: "integer", NotNull: true},
		{Name: "name", OriginalName: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", OriginalName: "bio", GoType: "sql.NullString", DBType: "text"},
	},
}

var getAuthorSQLite = &Query{
	Name:    "GetAuthor",
	Cmd:     ":one",
	Dialect: DialectSQLite,
	SQL:     "SELECT id, name, bio FROM authors WHERE id = ?1 LIMIT 1",
	Columns: []Column{
		{Name: "id", OriginalName: "id", GoType: "int64", DBType: "integer", NotNull: true},
		{Name: "name", OriginalName: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", OriginalName: "bio", GoType: "sql.NullString", DBType: "text"},
	},
	Args: []Arg{{Position: 1, GoType: "int64", DBType: "integer", NotNull: true}},
}

func TestBuilder_SQLite_WhereUsesQuestionMark(t *testing.T) {
	sql, args, err := Wrap(listAuthorsSQLite).
		Where("name", OpLike, "%foo%").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%foo%" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"name" LIKE ?1`) {
		t.Fatalf("expected SQLite-style ?1, got: %q", sql)
	}
	if strings.Contains(sql, "$1") {
		t.Fatalf("SQLite output should not contain $N placeholders: %q", sql)
	}
}

func TestBuilder_SQLite_RenumbersOverBaseArgs(t *testing.T) {
	sql, args, err := Wrap(getAuthorSQLite, int64(42)).
		Where("name", "=", "alice").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 || args[0] != int64(42) || args[1] != "alice" {
		t.Fatalf("args = %v", args)
	}
	// Original ?1 stays inside CTE; appended filter uses ?2.
	if !strings.Contains(sql, "WHERE id = ?1") {
		t.Fatalf("base ?1 should be preserved: %q", sql)
	}
	if !strings.Contains(sql, `"name" = ?2`) {
		t.Fatalf("expected ?2 for appended arg, got: %q", sql)
	}
}

func TestBuilder_SQLite_WhereExprRenumbers(t *testing.T) {
	sql, args, err := Wrap(listAuthorsSQLite).
		WhereExpr("name LIKE ? OR bio LIKE ?", "%a%", "%b%").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, "name LIKE ?1 OR bio LIKE ?2") {
		t.Fatalf("unexpected sql: %q", sql)
	}
}

func TestBuilder_SQLite_ILikeTranslatesToLike(t *testing.T) {
	sql, args, err := Wrap(listAuthorsSQLite).
		ApplyFilter(NewTextCol("name").ILike("%foo%")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%foo%" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"name" LIKE ?1`) {
		t.Fatalf("expected LIKE for SQLite, got: %q", sql)
	}
	if strings.Contains(sql, "ILIKE") {
		t.Fatalf("SQLite output should not contain ILIKE: %q", sql)
	}
}

func TestBuilder_InEmitsPortableInList(t *testing.T) {
	// Postgres dialect — `?` renumbers to `$N`, one slot per value.
	sql, args, err := Wrap(listAuthors).
		ApplyFilter(NewTextCol("name").In("ada", "alan", "grace")).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 3 || args[0] != "ada" || args[1] != "alan" || args[2] != "grace" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"name" IN ($1, $2, $3)`) {
		t.Fatalf("expected portable IN list, got: %q", sql)
	}
}

func TestBuilder_SQLite_InEmitsPortableInList(t *testing.T) {
	sql, args, err := Wrap(listAuthorsSQLite).
		ApplyFilter(NewIntCol("id").In(1, 2, 3)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 3 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sql, `"id" IN (?1, ?2, ?3)`) {
		t.Fatalf("expected portable IN list with SQLite placeholders, got: %q", sql)
	}
}
