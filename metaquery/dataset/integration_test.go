package dataset_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/dataset"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqsqlite"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/search"
	_ "modernc.org/sqlite"
)

// End-to-end: parse → compile → build SQL → execute on real SQLite → scan rows.
// This is the only test that runs the generated SQL through a database engine,
// catching anything string assertions can't (placeholder renumbering, CTE
// wrapping, LIKE ... ESCAPE, count queries, struct scanning).

type userRow struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Bio   string `db:"bio"`
	OrgID int64  `db:"org_id"`
	Score int64  `db:"score"`
}

func usersQuery() *metaquery.Query {
	return &metaquery.Query{
		Name:    "ListUsers",
		Dialect: metaquery.DialectSQLite,
		SQL:     "SELECT id, name, bio, org_id, score FROM users",
		Columns: []metaquery.Column{
			{Name: "id", GoType: "int64", DBType: "integer", NotNull: true},
			{Name: "name", GoType: "string", DBType: "text", NotNull: true},
			{Name: "bio", GoType: "string", DBType: "text", NotNull: true},
			{Name: "org_id", GoType: "int64", DBType: "integer", NotNull: true},
			{Name: "score", GoType: "int64", DBType: "integer", NotNull: true},
		},
	}
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	stmts := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, bio TEXT, org_id INTEGER, score INTEGER)`,
		`INSERT INTO users VALUES (1, 'Alice',        'admin user',      1,  90)`,
		`INSERT INTO users VALUES (2, 'Bob',          'regular user',    1,  60)`,
		`INSERT INTO users VALUES (3, 'Alice Cooper', 'musician',        2,  75)`,
		`INSERT INTO users VALUES (4, 'Carol',        'alice''s friend', 2, 100)`,
		`INSERT INTO users VALUES (5, '50% off',      'promo',           1,  50)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("setup %q: %v", s, err)
		}
	}
	return db
}

func run(t *testing.T, db *sql.DB, req dataset.Request, cfg dataset.Config) dataset.Response[userRow] {
	t.Helper()
	newB := func() *metaquery.Builder { return metaquery.Wrap(usersQuery()) }
	res, err := dataset.RunWithPartitionCount[userRow](context.Background(), newB, req, cfg,
		func(ctx context.Context, b *metaquery.Builder) (*metaquery.TypedResult[userRow], error) {
			return mqsqlite.Scan[userRow](ctx, db, b)
		},
		func(ctx context.Context, b *metaquery.Builder) (int64, error) {
			return mqsqlite.Count(ctx, db, b)
		},
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res
}

func names(rows []userRow) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Name
	}
	return out
}

func TestIntegration_GlobalSearch(t *testing.T) {
	db := openDB(t)
	// "alice" hits name OR bio (case-insensitive LIKE): Alice, Alice Cooper,
	// and Carol (bio "alice's friend").
	res := run(t, db, dataset.Request{Search: "alice", ShowCounts: true}, dataset.Config{})
	got := names(res.Data)
	if len(got) != 3 {
		t.Fatalf("want 3 rows, got %v", got)
	}
	if res.Count == nil || res.Count.InQuery != 3 {
		t.Fatalf("InQuery: %+v", res.Count)
	}
}

func TestIntegration_PartitionAndCounts(t *testing.T) {
	db := openDB(t)
	// Partition org_id=1 → {Alice, Bob, 50% off} (inPartition=3); + search
	// "alice" → just Alice (inQuery=1).
	res := run(t, db, dataset.Request{
		Partition:  "org_id:1",
		Search:     "alice",
		ShowCounts: true,
	}, dataset.Config{})
	if got := names(res.Data); len(got) != 1 || got[0] != "Alice" {
		t.Fatalf("data: %v", got)
	}
	if res.Count.InQuery != 1 || res.Count.InPartition != 3 {
		t.Fatalf("counts: %+v", res.Count)
	}
}

func TestIntegration_TargetedSearch(t *testing.T) {
	db := openDB(t)
	res := run(t, db, dataset.Request{Search: `name:"Alice Cooper"`}, dataset.Config{})
	if got := names(res.Data); len(got) != 1 || got[0] != "Alice Cooper" {
		t.Fatalf("data: %v", got)
	}
}

func TestIntegration_OrderingAndPagination(t *testing.T) {
	db := openDB(t)
	// All 5 rows, ordered by name DESC, page 0 size 2.
	res := run(t, db, dataset.Request{
		Ordering:   []dataset.Order{{Field: "name", Order: "DESC"}},
		Page:       0,
		PageSize:   2,
		ShowCounts: true,
	}, dataset.Config{Orderable: []string{"name"}})
	got := names(res.Data)
	// DESC by name: Carol, Bob, Alice Cooper, Alice, 50% off → page0/size2 = Carol, Bob.
	if len(got) != 2 || got[0] != "Carol" || got[1] != "Bob" {
		t.Fatalf("page0: %v", got)
	}
	if res.Count.InQuery != 5 {
		t.Fatalf("InQuery: %+v", res.Count)
	}
}

func TestIntegration_LikeWildcardEscaped(t *testing.T) {
	db := openDB(t)
	// Searching "50%" must match the literal "50% off" row, NOT treat % as a
	// wildcard (which would match every row). Exercises the LIKE ... ESCAPE path.
	res := run(t, db, dataset.Request{Search: "50%"}, dataset.Config{})
	if got := names(res.Data); len(got) != 1 || got[0] != "50% off" {
		t.Fatalf("escaped %% search should match only '50%% off', got %v", got)
	}
}

func TestIntegration_ValueComparison(t *testing.T) {
	db := openDB(t)
	// score >= 90 → Alice(90), Carol(100). Numeric defaults are operator-aware.
	res := run(t, db, dataset.Request{Search: "score:>=90"}, dataset.Config{})
	got := names(res.Data)
	if len(got) != 2 {
		t.Fatalf("score:>=90 → %v", got)
	}
}

func TestIntegration_ValueRange(t *testing.T) {
	db := openDB(t)
	// score BETWEEN 60 AND 90 → Bob(60), Alice Cooper(75), Alice(90).
	res := run(t, db, dataset.Request{Search: "score:60..90"}, dataset.Config{})
	if got := names(res.Data); len(got) != 3 {
		t.Fatalf("score:60..90 → %v", got)
	}
}

func TestIntegration_WildcardPrefix(t *testing.T) {
	db := openDB(t)
	// name LIKE 'Alic%' → Alice, Alice Cooper (not Carol/Bob).
	cfg := dataset.Config{}
	cfg.Config.Fields = map[string]search.Field{"name": {Search: search.WildcardFn()}}
	res := run(t, db, dataset.Request{Search: "name:Alic*"}, cfg)
	got := names(res.Data)
	if len(got) != 2 {
		t.Fatalf("name:Alic* → %v", got)
	}
	for _, n := range got {
		if n != "Alice" && n != "Alice Cooper" {
			t.Fatalf("unexpected match: %v", got)
		}
	}
}

func TestIntegration_TargetableOperatorsUnderAllowlist(t *testing.T) {
	db := openDB(t)
	// Mirrors redline's region config: name free-text, score targeted-only with
	// operators. Under the hard allowlist, score:>=90 still works (Targetable),
	// while an unlisted column stays disabled.
	cfg := dataset.Config{Searchable: []string{"name"}, Targetable: []string{"score"}}
	res := run(t, db, dataset.Request{Search: "score:>=90"}, cfg)
	if got := names(res.Data); len(got) != 2 {
		t.Fatalf("score:>=90 under allowlist → %v", got)
	}
	// org_id not listed → not targetable → falls back to global search of "1".
	res = run(t, db, dataset.Request{Search: "org_id:1"}, cfg)
	if len(res.Data) != 0 {
		t.Fatalf("org_id should be disabled under allowlist: %v", names(res.Data))
	}
}

func TestIntegration_HardAllowlistNotTargetable(t *testing.T) {
	db := openDB(t)
	// org_id excluded from the allowlist: org_id:1 must not filter by org_id;
	// it falls back to a global value search for "1" (matches nothing here).
	res := run(t, db, dataset.Request{Search: "org_id:1"}, dataset.Config{Searchable: []string{"name"}})
	if len(res.Data) != 0 {
		t.Fatalf("org_id should not be targetable under a hard allowlist: %v", names(res.Data))
	}
}
