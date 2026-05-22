// demo is a CLI for exercising sqlc-go-codegen-metaquery against a local
// SQLite database. Mirrors examples/pgx but uses database/sql + mqsqlite, so
// it runs without Docker — the DB file is created on first run.
//
// Usage: see README.md.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/examples/sqlite/db"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqsqlite"

	_ "modernc.org/sqlite"
)

const defaultDSN = "file:demo.db?_pragma=foreign_keys(1)"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]

	ctx := context.Background()
	switch cmd {
	case "init":
		mustRun(runInit(ctx, args))
	case "seed":
		mustRun(runSeed(ctx, args))
	case "list":
		mustRun(runList(ctx, args))
	case "counts":
		mustRun(runCounts(ctx, args))
	case "search":
		mustRun(runSearch(ctx, args))
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `demo — sqlc-go-codegen-metaquery example CLI (SQLite)

Commands:
  init                        Apply migrations/001_schema.sql to the DB.
  seed                        Populate sample authors + posts.
  list   [flags]              List authors via mqsqlite.Run (untyped Result).
  list --typed [flags]        List authors via mqsqlite.Scan[db.Author] (typed).
  counts [flags]              Post counts per author (aggregation via Scan[T]).
  search <needle>             List authors whose name matches, using typed Col refs.

Common flags:
  --db      sqlite DSN (default "`+defaultDSN+`")
  --where   JSON array of metaquery.Filter, e.g. '[{"column":"name","op":"LIKE","value":"%a%"}]'
  --order   JSON array of metaquery.OrderBy, e.g. '[{"column":"name","dir":"ASC"}]'
  --page    0-indexed page (default 0)
  --size    Page size (default 0 = unlimited)
  --total   Compute total row count (default true when --size > 0)
`)
}

type commonFlags struct {
	dsn       string
	whereJSON string
	orderJSON string
	page      int
	size      int
	total     bool
	typed     bool
}

func bindCommon(fs *flag.FlagSet) *commonFlags {
	c := &commonFlags{}
	fs.StringVar(&c.dsn, "db", defaultDSN, "sqlite DSN")
	fs.StringVar(&c.whereJSON, "where", "", "JSON array of metaquery.Filter")
	fs.StringVar(&c.orderJSON, "order", "", "JSON array of metaquery.OrderBy")
	fs.IntVar(&c.page, "page", 0, "0-indexed page")
	fs.IntVar(&c.size, "size", 0, "page size (0 = unlimited)")
	fs.BoolVar(&c.total, "total", true, "compute total row count")
	fs.BoolVar(&c.typed, "typed", false, "use typed Scan[T] instead of untyped Run")
	return c
}

func (c *commonFlags) applyTo(b *metaquery.Builder) error {
	if c.whereJSON != "" {
		var fs []metaquery.Filter
		if err := json.Unmarshal([]byte(c.whereJSON), &fs); err != nil {
			return fmt.Errorf("--where: %w", err)
		}
		b.ApplyFilters(fs)
	}
	if c.orderJSON != "" {
		var os []metaquery.OrderBy
		if err := json.Unmarshal([]byte(c.orderJSON), &os); err != nil {
			return fmt.Errorf("--order: %w", err)
		}
		b.ApplyOrders(os)
	}
	if c.size > 0 {
		b.ApplyPagination(metaquery.PageRequest{Page: c.page, Size: c.size, Total: c.total})
	}
	return nil
}

func open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

// runInit applies the schema. Idempotent: re-running on an initialised DB
// is a no-op because the CREATE statements use IF NOT EXISTS — wait, ours
// don't. We just swallow "already exists" errors instead.
func runInit(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	dsn := fs.String("db", defaultDSN, "sqlite DSN")
	fs.Parse(args)

	schema, err := os.ReadFile("migrations/001_schema.sql")
	if err != nil {
		return err
	}
	conn, err := open(*dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, string(schema)); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "schema applied")
	return nil
}

func runSeed(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	dsn := fs.String("db", defaultDSN, "sqlite DSN")
	fs.Parse(args)

	conn, err := open(*dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	q := db.New(conn)
	if err := q.DeleteAllPosts(ctx); err != nil {
		return err
	}
	if err := q.DeleteAllAuthors(ctx); err != nil {
		return err
	}

	authors := []db.CreateAuthorParams{
		{Name: "Ada Lovelace", Bio: sql.NullString{String: "First programmer.", Valid: true}},
		{Name: "Alan Turing", Bio: sql.NullString{String: "Father of CS.", Valid: true}},
		{Name: "Barbara Liskov"},
		{Name: "Bjarne Stroustrup"},
		{Name: "Grace Hopper", Bio: sql.NullString{String: "Amazing Grace.", Valid: true}},
	}
	postsPerAuthor := map[string]int{
		"Ada Lovelace":      3,
		"Alan Turing":       5,
		"Barbara Liskov":    2,
		"Bjarne Stroustrup": 4,
		"Grace Hopper":      1,
	}
	for _, ap := range authors {
		a, err := q.CreateAuthor(ctx, ap)
		if err != nil {
			return err
		}
		n := postsPerAuthor[a.Name]
		for i := range n {
			if _, err := q.CreatePost(ctx, db.CreatePostParams{
				AuthorID: a.ID,
				Title:    fmt.Sprintf("%s — post %d", a.Name, i+1),
				Body:     sql.NullString{String: "Lorem ipsum.", Valid: true},
				Views:    int64(100 * (i + 1)),
			}); err != nil {
				return err
			}
		}
	}
	fmt.Fprintln(os.Stderr, "seeded")
	return nil
}

func runList(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	c := bindCommon(fs)
	fs.Parse(args)

	conn, err := open(c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	b := db.WrapListAuthors()
	if err := c.applyTo(b); err != nil {
		return err
	}

	if c.typed {
		res, err := mqsqlite.Scan[db.Author](ctx, conn, b)
		if err != nil {
			return err
		}
		return writeJSON(res)
	}
	res, err := mqsqlite.Run(ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

// AuthorPostCount is the user-defined row type for the aggregation demo.
// `db` tags are how mqsqlite.Scan[T] maps columns to fields.
type AuthorPostCount struct {
	AuthorID int64 `db:"author_id" json:"author_id"`
	Total    int64 `db:"total"     json:"total"`
}

func runCounts(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("counts", flag.ExitOnError)
	c := bindCommon(fs)
	fs.Parse(args)

	conn, err := open(c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	b := db.WrapListPosts().
		GroupBy("author_id").
		Count("total")
	if err := c.applyTo(b); err != nil {
		return err
	}

	res, err := mqsqlite.Scan[AuthorPostCount](ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

// runSearch demonstrates the typed col-ref API. SQLite's LIKE is
// case-insensitive for ASCII by default, so we use .Like instead of the
// Postgres-specific .ILike.
func runSearch(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	c := bindCommon(fs)
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("search: expected a needle argument")
	}
	needle := rest[0]

	conn, err := open(c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close()

	// .ILike works on SQLite too: the Builder translates ILIKE to LIKE at
	// Build time when Dialect is SQLite. SQLite's LIKE is case-insensitive
	// for ASCII by default, matching the typical use of ILIKE.
	b := db.WrapListAuthors().
		ApplyFilter(db.ListAuthorsCols.Name.ILike("%" + needle + "%")).
		ApplyOrder(db.ListAuthorsCols.Name.Asc())
	if c.size > 0 {
		b.ApplyPagination(metaquery.PageRequest{Page: c.page, Size: c.size, Total: c.total})
	}

	res, err := mqsqlite.Scan[db.Author](ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func mustRun(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
