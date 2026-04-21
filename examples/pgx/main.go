// demo is a CLI for exercising sqlc-go-codegen-metaquery against a live
// PostgreSQL instance. It exposes three subcommands (seed, list, counts) that
// demonstrate the untyped Run path, typed Scan[T] path, and aggregation path.
// All non-seed commands output JSON.
//
// Usage: see README.md.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/iodesystems/sqlc-go-codegen-metaquery/examples/pgx/db"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
	"github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const defaultDSN = "postgres://demo:demo@localhost:5544/demo?sslmode=disable"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]

	ctx := context.Background()
	switch cmd {
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
	fmt.Fprint(os.Stderr, `demo — sqlc-go-codegen-metaquery example CLI

Commands:
  seed                       Populate sample authors + posts.
  list   [flags]             List authors via mqpgx.Run (untyped Result).
  list --typed [flags]       List authors via mqpgx.Scan[db.Author] (typed).
  counts [flags]             Post counts per author (aggregation via Scan[T]).
  search <needle>            List authors whose name matches, using typed Col refs.

Common flags:
  --db      DATABASE_URL (default "`+defaultDSN+`")
  --where   JSON array of metaquery.Filter, e.g. '[{"column":"name","op":"ILIKE","value":"%a%"}]'
  --order   JSON array of metaquery.OrderBy, e.g. '[{"column":"name","dir":"ASC"}]'
  --page    0-indexed page (default 0)
  --size    Page size (default 0 = unlimited)
  --total   Compute total row count (default true when --size > 0)

Output is a JSON envelope: {"data":[...], "meta":{"columns":..., "pagination":..., "where":..., ...}}
`)
}

// ---- common flag parsing ----

type commonFlags struct {
	dsn      string
	whereJSON string
	orderJSON string
	page      int
	size      int
	total     bool
	typed     bool
}

func bindCommon(fs *flag.FlagSet) *commonFlags {
	c := &commonFlags{}
	fs.StringVar(&c.dsn, "db", defaultDSN, "DATABASE_URL")
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

// ---- subcommands ----

func runSeed(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	dsn := fs.String("db", defaultDSN, "DATABASE_URL")
	fs.Parse(args)

	conn, err := pgx.Connect(ctx, *dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	q := db.New(conn)
	if err := q.TruncateAll(ctx); err != nil {
		return err
	}

	authors := []db.CreateAuthorParams{
		{Name: "Ada Lovelace", Bio: pgtype.Text{String: "First programmer.", Valid: true}},
		{Name: "Alan Turing", Bio: pgtype.Text{String: "Father of CS.", Valid: true}},
		{Name: "Barbara Liskov", Bio: pgtype.Text{}},
		{Name: "Bjarne Stroustrup"},
		{Name: "Grace Hopper", Bio: pgtype.Text{String: "Amazing Grace.", Valid: true}},
	}
	type seeded struct {
		id    int64
		posts int
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
		for i := 0; i < n; i++ {
			if _, err := q.CreatePost(ctx, db.CreatePostParams{
				AuthorID: a.ID,
				Title:    fmt.Sprintf("%s — post %d", a.Name, i+1),
				Body:     pgtype.Text{String: "Lorem ipsum.", Valid: true},
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

	conn, err := pgx.Connect(ctx, c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Typed wrapper: no positional args to bind, so no type parameters needed;
	// callers of queries with args get compile-time checks (e.g. db.WrapGetAuthor(42)).
	b := db.WrapListAuthors()
	if err := c.applyTo(b); err != nil {
		return err
	}

	if c.typed {
		res, err := mqpgx.Scan[db.Author](ctx, conn, b)
		if err != nil {
			return err
		}
		return writeJSON(res)
	}
	res, err := mqpgx.Run(ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

// AuthorPostCount is the user-defined row type for the aggregation demo.
// `db` tags are how mqpgx.Scan[T] maps columns to fields.
type AuthorPostCount struct {
	AuthorID int64  `db:"author_id" json:"author_id"`
	Total    int64  `db:"total"     json:"total"`
}

func runCounts(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("counts", flag.ExitOnError)
	c := bindCommon(fs)
	fs.Parse(args)

	conn, err := pgx.Connect(ctx, c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Wrap ListPosts via the typed constructor (no args here); add group/agg.
	b := db.WrapListPosts().
		GroupBy("author_id").
		Count("total")
	if err := c.applyTo(b); err != nil {
		return err
	}

	res, err := mqpgx.Scan[AuthorPostCount](ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

// runSearch demonstrates the typed col-ref API: compile-time column names,
// compile-time op/value types. Contrast with `list --where ...` which uses
// JSON-parsed filters — both end up as the same metaquery.Filter internally.
func runSearch(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	c := bindCommon(fs)
	fs.Parse(args)

	rest := fs.Args()
	if len(rest) == 0 {
		return fmt.Errorf("search: expected a needle argument")
	}
	needle := rest[0]

	conn, err := pgx.Connect(ctx, c.dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// Typed cols — db.ListAuthorsCols.Name is a TextCol; .ILike() takes string.
	// `db.ListAuthorsCols.Name.ILike(5)` would be a compile error.
	b := db.WrapListAuthors().
		ApplyFilter(db.ListAuthorsCols.Name.ILike("%" + needle + "%")).
		ApplyOrder(db.ListAuthorsCols.Name.Asc())
	if c.size > 0 {
		b.ApplyPagination(metaquery.PageRequest{Page: c.page, Size: c.size, Total: c.total})
	}

	res, err := mqpgx.Scan[db.Author](ctx, conn, b)
	if err != nil {
		return err
	}
	return writeJSON(res)
}

// ---- helpers ----

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func mustRun(err error) {
	if err == nil {
		return
	}
	var perr *pgconn.PgError
	if errors.As(err, &perr) {
		fmt.Fprintf(os.Stderr, "pg error: %s (code %s)\n", perr.Message, perr.Code)
	} else {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
	os.Exit(1)
}
