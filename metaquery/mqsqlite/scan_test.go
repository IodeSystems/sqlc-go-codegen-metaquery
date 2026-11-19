package mqsqlite_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqsqlite"
	_ "modernc.org/sqlite"
)

// End-to-end proof of the SQLite adapter against a real in-memory SQLite
// (modernc.org/sqlite, pure Go): the builder's SQLite dialect emits ?N
// placeholders (never $N) and translates an ILIKE filter to LIKE, the
// wrapped query executes and rows map into T, and WithTotal issues the
// COUNT follow-up. String-only tests in metaquery pin the SQL shape; this
// test pins it reaching a database engine and executing.

type authorRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	Bio  string `db:"bio"`
}

var listAuthors = &metaquery.Query{
	Name:    "ListAuthors",
	Cmd:     ":many",
	Dialect: metaquery.DialectSQLite,
	SQL:     "SELECT id, name, bio FROM authors ORDER BY name",
	Columns: []metaquery.Column{
		{Name: "id", GoType: "int64", DBType: "integer", NotNull: true},
		{Name: "name", GoType: "string", DBType: "text", NotNull: true},
		{Name: "bio", GoType: "string", DBType: "text", NotNull: true},
	},
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	stmts := []string{
		`CREATE TABLE authors (id INTEGER PRIMARY KEY, name TEXT, bio TEXT)`,
		`INSERT INTO authors VALUES (1, 'Alice',  'alice in the bio')`,
		`INSERT INTO authors VALUES (2, 'Bob',   'bob row')`,
		`INSERT INTO authors VALUES (3, 'Carol', 'alice fan')`,
		`INSERT INTO authors VALUES (4, 'Dave',  'nothing')`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

// The built SQL for an ILIKE filter on the SQLite dialect is LIKE with a
// numbered ? placeholder, and contains no ILIKE and no $N.
func TestSQLite_ILikeTranslatesToLikeAndQuestionN(t *testing.T) {
	b := metaquery.Wrap(listAuthors).ApplyFilter(metaquery.NewTextCol("name").ILike("%alice%"))
	sqlText, args, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 1 || args[0] != "%alice%" {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sqlText, `"name" LIKE ?1`) {
		t.Fatalf("expected LIKE ?1 for SQLite ILIKE, got: %q", sqlText)
	}
	if strings.Contains(sqlText, "ILIKE") {
		t.Fatalf("SQLite output should not contain ILIKE: %q", sqlText)
	}
	if strings.Contains(sqlText, "$") {
		t.Fatalf("SQLite output should not contain $N placeholders: %q", sqlText)
	}
}

// The ILIKE-built filter executes as a substring match and Scan[T] maps the
// rows: "alice" matches Alice and the bio-mentioning Carol, case
// insensitively in spirit (ASCII LIKE), with the args bound positionally.
func TestSQLite_ScanWithILikeFilter(t *testing.T) {
	db := openDB(t)
	b := metaquery.Wrap(listAuthors).ApplyFilter(metaquery.NewTextCol("name").ILike("%alice%"))
	res, err := mqsqlite.Scan[authorRow](context.Background(), db, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Data) != 1 {
		t.Fatalf("name LIKE %%alice%% matched %d rows, want 1 (Alice): %+v", len(res.Data), res.Data)
	}
	if res.Data[0].Name != "Alice" || res.Data[0].ID != 1 {
		t.Fatalf("unexpected row: %+v", res.Data[0])
	}

	// The filter runs on bio too, proving the translation, not just the
	// operator name: Carol's bio mentions alice.
	bio := metaquery.Wrap(listAuthors).ApplyFilter(metaquery.NewTextCol("bio").ILike("%alice%"))
	res2, err := mqsqlite.Scan[authorRow](context.Background(), db, bio)
	if err != nil {
		t.Fatal(err)
	}
	if len(res2.Data) != 2 {
		t.Fatalf("bio LIKE %%alice%% matched %d rows, want 2 (Alice, Carol): %+v", len(res2.Data), res2.Data)
	}
}

// A second ?N placeholder is allocated after the first: base-arg renumbering
// is what makes ?N (not bare ?) necessary, so prove two appended filters get
// ?1 and ?2 and execute.
func TestSQLite_MultipleFiltersRenumberPlaceholders(t *testing.T) {
	b := metaquery.Wrap(listAuthors).
		ApplyFilter(metaquery.NewTextCol("name").ILike("%a%")).
		ApplyFilter(metaquery.NewTextCol("bio").ILike("%user%"))
	sqlText, args, err := b.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 2 {
		t.Fatalf("args = %v", args)
	}
	if !strings.Contains(sqlText, `LIKE ?1`) || !strings.Contains(sqlText, `LIKE ?2`) {
		t.Fatalf("expected ?1 and ?2, got: %q", sqlText)
	}

	db := openDB(t)
	res, err := mqsqlite.Scan[authorRow](context.Background(), db, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Data) != 0 {
		t.Fatalf("no row has both name and bio matching, got %+v", res.Data)
	}
}

// WithTotal issues the COUNT follow-up and populates Pagination.Total.
func TestSQLite_WantsTotalCountFollowUp(t *testing.T) {
	db := openDB(t)
	b := metaquery.Wrap(listAuthors).
		ApplyFilter(metaquery.NewTextCol("name").ILike("%a%")).
		WithTotal()
	res, err := mqsqlite.Run(context.Background(), db, b)
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Pagination == nil {
		t.Fatal("expected Pagination to be populated")
	}
	// Alice, Bob (bio has no a... name LIKE %a%: Alice, Bob(no), Carol(no), Dave)
	// name LIKE %a% → Alice, Bob? no. Alice (name has a), Bob no, Carol (a in Carol? C-a-r-o-l yes), Dave (a yes)
	if res.Meta.Pagination.Total != 3 {
		t.Fatalf("total = %d, want 3 (Alice, Carol, Dave): %+v", res.Meta.Pagination.Total, res.Meta.Pagination)
	}
	if len(res.Data) != 3 {
		t.Fatalf("data rows = %d, want 3", len(res.Data))
	}
}

// Count runs the BuildCount SQL standalone with the same placeholder form.
func TestSQLite_CountUsesQuestionN(t *testing.T) {
	db := openDB(t)
	b := metaquery.Wrap(listAuthors).ApplyFilter(metaquery.NewTextCol("name").ILike("%a%"))
	total, err := mqsqlite.Count(context.Background(), db, b)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 {
		t.Fatalf("count = %d, want 3", total)
	}
}
