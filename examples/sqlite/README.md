# examples/sqlite — runnable demo for sqlc-go-codegen-metaquery

End-to-end example: SQLite (via the pure-Go [`modernc.org/sqlite`][modernc]
driver — no CGO, no external services), sqlc-generated queries + metadata,
and a tiny CLI that mirrors `examples/pgx` against SQLite.

[modernc]: https://pkg.go.dev/modernc.org/sqlite

## Prerequisites

- Go 1.21+
- [`sqlc`](https://sqlc.dev) on `$PATH`
- Parent repo has built the plugin binary: `make -C ../.. bin/sqlc-go-codegen-metaquery`

## Full walk-through

```sh
make demo
```

That runs, in order: regenerates `db/*.go`, drops and recreates `demo.db`,
seeds sample data, and calls each subcommand.

Step-by-step equivalents:

```sh
make generate     # regenerate db/*.go via the plugin
make init         # fresh demo.db + apply schema
make seed         # seed 5 authors, 15 posts
make list         # list authors (untyped Result)
make list-typed   # list authors (TypedResult[db.Author])
make counts       # post counts grouped by author_id (TypedResult[AuthorPostCount])
make clean        # rm bin/ demo.db
```

## SQLite-specific notes

- **Engine.** `sqlc.yaml` sets `engine: sqlite` and `sql_package: database/sql`.
  The plugin emits `Dialect: metaquery.DialectSQLite` on every `Meta<Name>`,
  which causes the builder to emit `?N` placeholders (Postgres' `$N` is not
  the conventional placeholder form for SQLite via `database/sql`).
- **Driver.** This demo uses `modernc.org/sqlite` (pure Go). Any
  `database/sql` SQLite driver works — `mattn/go-sqlite3`, libsql, etc. —
  swap the import and the `sql.Open("sqlite", ...)` driver name.
- **`ILIKE`.** SQLite has no `ILIKE`, but the Builder translates it to
  `LIKE` at Build time when Dialect is SQLite. SQLite's `LIKE` is
  case-insensitive for ASCII by default, so `.ILike(...)` on typed text cols
  works the same way as in the pgx demo. See `runSearch` in `main.go`.
- **`In(...)`.** `TextCol.In` and `IntCol.In` now emit portable
  `IN (?, ?, ...)` SQL (one placeholder per value), so they work on both
  Postgres and SQLite.

## CLI reference

```
demo init                       Apply migrations/001_schema.sql to the DB.
demo seed                       Populate sample authors + posts.
demo list   [flags]             List authors via mqsqlite.Run (untyped Result).
demo list --typed [flags]       List authors via mqsqlite.Scan[db.Author] (typed).
demo counts [flags]             Post counts per author (aggregation via Scan[T]).
demo search <needle>            List authors whose name matches, using typed Col refs.
```

Common flags:

| Flag      | Meaning                                                               |
| --------- | --------------------------------------------------------------------- |
| `--db`    | sqlite DSN. Default `file:demo.db?_pragma=foreign_keys(1)`            |
| `--where` | JSON array of `metaquery.Filter`                                      |
| `--order` | JSON array of `metaquery.OrderBy`                                     |
| `--page`  | 0-indexed page                                                        |
| `--size`  | Page size (`0` = unlimited)                                           |
| `--total` | Also run a `COUNT(*)` to populate `meta.pagination.total` (default `true` when `--size > 0`) |
