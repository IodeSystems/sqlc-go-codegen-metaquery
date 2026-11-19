package mqpgx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// The pgx adapter is proven with a fake Queryer: it records every SQL text
// and argument slice the adapter issues and answers from hand-built rows.
// That pins the adapter's contract — $N placeholders reaching the wire,
// RowToStructByName mapping, and the WithTotal COUNT follow-up — without a
// live Postgres. A real-server run is out of reach in a dev session; the
// fake proves the shape the server would parse.

type authorRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Bio  string `db:"bio"`
}

var listAuthors = &metaquery.Query{
	Name: "ListAuthors",
	Cmd:  ":many",
	SQL:  "SELECT id, name, bio FROM authors ORDER BY name",
	Columns: []metaquery.Column{
		{Name: "id", OriginalName: "id", GoType: "int64", DBType: "int8", NotNull: true},
		{Name: "name", OriginalName: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", OriginalName: "bio", GoType: "string", DBType: "text", NotNull: true},
	},
}

func rowsData(ids []int64) [][]any {
	rows := make([][]any, len(ids))
	for i, id := range ids {
		rows[i] = []any{id, "author-" + string(rune('a'+i%26)), "bio"}
	}
	return rows
}

// fakeRows is a hand-built pgx.Rows: three columns, values returned by
// Values(). FieldDescriptions() carries the names RowToStructByName keys
// off.
type fakeRows struct {
	data  [][]any
	names []string
	i     int
}

func (r *fakeRows) Close() {}

func (r *fakeRows) Err() error { return nil }

func (r *fakeRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription {
	fds := make([]pgconn.FieldDescription, len(r.names))
	for i, n := range r.names {
		fds[i] = pgconn.FieldDescription{Name: n}
	}
	return fds
}

func (r *fakeRows) Next() bool {
	r.i++
	return r.i < len(r.data)
}

func (r *fakeRows) Scan(dest ...any) error {
	if r.i < 0 || r.i >= len(r.data) {
		return errors.New("fakeRows: Scan without Next")
	}
	for i, d := range dest {
		if d == nil {
			continue
		}
		v := r.data[r.i][i]
		switch p := d.(type) {
		case *any:
			*p = v
		case *int64:
			*p = v.(int64)
		case *string:
			*p = v.(string)
		default:
			return fmt.Errorf("fakeRows: unsupported destination %T", d)
		}
	}
	return nil
}

func (r *fakeRows) Values() ([]any, error) {
	if r.i < 0 || r.i >= len(r.data) {
		return nil, errors.New("fakeRows: Values without Next")
	}
	return r.data[r.i], nil
}

func (r *fakeRows) RawValues() [][]byte { return nil }

func (r *fakeRows) Conn() *pgx.Conn { return nil }

func (r *fakeRows) TypeMap() *pgtype.Map { return pgtype.NewMap() }

// fakeQueryer records each call and answers from the data seeded in it:
// data for the wrapped query, total for the COUNT follow-up.
type fakeQueryer struct {
	calls []call
	data  [][]any
	total int64
}

type call struct {
	sql  string
	args []any
}

func (f *fakeQueryer) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	f.calls = append(f.calls, call{sql: sql, args: args})
	if strings.Contains(sql, "count(*)") {
		return &fakeRows{i: -1, data: [][]any{{f.total}}, names: []string{"count"}}, nil
	}
	return &fakeRows{i: -1, data: f.data, names: []string{"id", "name", "bio"}}, nil
}

func (f *fakeQueryer) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	f.calls = append(f.calls, call{sql: sql, args: args})
	return &fakeRow{v: f.total}
}

type fakeRow struct{ v any }

func (r *fakeRow) Scan(dest ...any) error {
	if len(dest) != 1 {
		return errors.New("fakeRow: one destination expected")
	}
	if p, ok := dest[0].(*int64); ok {
		*p = r.v.(int64)
		return nil
	}
	return errors.New("fakeRow: *int64 destination expected")
}

func TestRun_UsesDollarPlaceholders(t *testing.T) {
	f := &fakeQueryer{data: rowsData([]int64{1, 2, 3})}
	b := metaquery.Wrap(listAuthors).
		Where("name", metaquery.OpILike, "%a%").
		OrderBy("name", metaquery.Asc).
		Limit(50).Offset(10)

	res, err := Run(context.Background(), f, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("expected 1 query, got %d", len(f.calls))
	}
	sql, args := f.calls[0].sql, f.calls[0].args
	// The adapter hands the builder's $N text to pgx untouched.
	if !strings.Contains(sql, `"name" ILIKE $1`) {
		t.Fatalf("missing $1 placeholder in %q", sql)
	}
	if !strings.Contains(sql, "LIMIT 50 OFFSET 10") {
		t.Fatalf("missing LIMIT/OFFSET in %q", sql)
	}
	if !strings.Contains(sql, "WITH __q AS (") || !strings.Contains(sql, listAuthors.SQL) {
		t.Fatalf("original SQL must survive verbatim in the CTE: %q", sql)
	}
	if len(args) != 1 || args[0] != "%a%" {
		t.Fatalf("args = %v", args)
	}
	if len(res.Data) != 3 {
		t.Fatalf("want 3 rows, got %d", len(res.Data))
	}
	if got := res.Meta.Where; len(got) != 1 || got[0].Column != "name" {
		t.Fatalf("meta.Where = %+v", got)
	}
}

func TestScan_TypedRows(t *testing.T) {
	f := &fakeQueryer{data: rowsData([]int64{7, 8})}
	b := metaquery.Wrap(listAuthors).Where("id", metaquery.OpGt, int64(5))

	res, err := Scan[authorRow](context.Background(), f, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("expected 1 query, got %d", len(f.calls))
	}
	if !strings.Contains(f.calls[0].sql, `"id" > $1`) {
		t.Fatalf("missing $1 in %q", f.calls[0].sql)
	}
	if len(res.Data) != 2 || res.Data[0].ID != 7 || res.Data[1].ID != 8 {
		t.Fatalf("typed rows = %+v", res.Data)
	}
	if res.Data[0].Name != "author-a" || res.Data[0].Bio != "bio" {
		t.Fatalf("column mapping wrong: %+v", res.Data[0])
	}
}

func TestScanOne(t *testing.T) {
	f := &fakeQueryer{data: rowsData([]int64{42})}
	row, meta, err := ScanOne[authorRow](context.Background(), f, metaquery.Wrap(listAuthors))
	if err != nil {
		t.Fatal(err)
	}
	if row.ID != 42 {
		t.Fatalf("row = %+v", row)
	}
	if got := len(meta.Columns); got != 3 {
		t.Fatalf("meta.Columns = %d", got)
	}
}

func TestRun_WithTotalRunsCountFollowUp(t *testing.T) {
	f := &fakeQueryer{data: rowsData([]int64{1, 2, 3}), total: 42}
	b := metaquery.Wrap(listAuthors).
		Where("name", metaquery.OpILike, "%a%").
		ApplyPagination(metaquery.PageRequest{Page: 0, Size: 2, Total: true})

	res, err := Run(context.Background(), f, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.calls) != 2 {
		t.Fatalf("expected data + count queries, got %d", len(f.calls))
	}
	data, count := f.calls[0], f.calls[1]
	if !strings.Contains(data.sql, "LIMIT 2") {
		t.Fatalf("data query missing LIMIT: %q", data.sql)
	}
	if !strings.Contains(count.sql, "SELECT count(*)") {
		t.Fatalf("count query = %q", count.sql)
	}
	// The count reuses the same base args and placeholder numbering.
	if !strings.Contains(count.sql, `"name" ILIKE $1`) || len(count.args) != 1 || count.args[0] != "%a%" {
		t.Fatalf("count call = %+v", count)
	}
	if res.Meta.Pagination == nil || res.Meta.Pagination.Total != 42 || res.Meta.Pagination.Limit != 2 {
		t.Fatalf("pagination = %+v", res.Meta.Pagination)
	}
	// The fake ignores LIMIT (it applies no SQL), so row count comes from the
	// seed; the LIMIT itself is pinned in the data query text above.
	if got := len(res.Data); got != 3 {
		t.Fatalf("want 3 rows, got %d", got)
	}
}

func TestCountStandalone(t *testing.T) {
	f := &fakeQueryer{total: 5}
	n, err := Count(context.Background(), f, metaquery.Wrap(listAuthors))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("count = %d", n)
	}
	if len(f.calls) != 1 || !strings.Contains(f.calls[0].sql, "SELECT count(*)") {
		t.Fatalf("calls = %+v", f.calls)
	}
}
