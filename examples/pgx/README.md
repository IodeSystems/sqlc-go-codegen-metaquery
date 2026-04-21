# examples/pgx — runnable demo for sqlc-go-codegen-metaquery

End-to-end example: PostgreSQL 17 in Docker, sqlc-generated queries +
metadata, and a tiny CLI that shows three shapes of output from the
`metaquery` runtime.

## Prerequisites

- Docker (with the `compose` plugin)
- Go 1.21+
- [`sqlc`](https://sqlc.dev) on `$PATH`
- Parent repo has built the plugin binary: `make -C ../.. bin/sqlc-go-codegen-metaquery`

## Full walk-through

```sh
make demo
```

That runs, in order: `docker compose up -d`, waits for pg, `sqlc generate`,
builds the CLI, seeds sample data, then calls each subcommand.

Step-by-step equivalents:

```sh
make up           # start pg17
make generate     # regenerate db/*.go via the plugin
make build        # build ./bin/demo
make seed         # truncate + seed 5 authors, 15 posts
make list         # list authors (untyped Result)
make list-typed   # list authors (TypedResult[db.Author])
make counts       # post counts grouped by author_id (TypedResult[AuthorPostCount])
make down         # stop + wipe volume
```

## Controlling what's generated — `emit_metaquery`

Metaquery emission is a ladder; each level adds to the one below it:

| Level  | Emits per query                                       | When to pick it |
| ------ | ----------------------------------------------------- | --------------- |
| `off`  | nothing                                               | Query is only called directly (`q.Foo(ctx, ...)`); no builder/reflection use. |
| `meta` | `MetaFoo metaquery.Query`                             | You want runtime metadata (columns, SQL text) without the builder ergonomics — e.g. you're wiring your own introspection. |
| `wrap` | `MetaFoo` + `WrapFoo(args...) *metaquery.Builder`     | You want compile-time-checked wrappers but filter against column names as strings. |
| `cols` | `MetaFoo` + `WrapFoo(...)` + `FooCols` typed col refs | Default. Everything: typed wrappers *and* typed column refs. |

### Global default — `sqlc.yaml`

```yaml
codegen:
- plugin: metaquery
  out: db
  options:
    package: db
    sql_package: pgx/v5
    emit_metaquery: cols   # default; off|meta|wrap|cols
```

### Per-query override — query comment

Put `-- metaquery: <level>` in the query's leading comment block. It overrides
the global default for that query only.

```sql
-- name: TruncateAll :exec
-- metaquery: off
-- (TruncateAll is only invoked from seed code; no builder/metadata needed.)
TRUNCATE authors RESTART IDENTITY CASCADE;
```

In this example repo, `TruncateAll` is excluded from emission entirely — the
generated `db/query.sql.metaquery.go` has no `MetaTruncateAll`. All other
queries stay at the default `cols` level.

### How to choose

- **Using filters/pagination/aggregation from Go code?** Leave the default
  (`cols`). The extra generated code is small (~20 LOC/query) and gives you
  compile-time safety.
- **Only using the `Meta*` vars to drive something custom** (schema export,
  generic API handler, introspection UI)? Set `emit_metaquery: meta` globally
  — halves the generated volume.
- **Forking a codebase that imports this plugin but doesn't need metaquery?**
  Set `emit_metaquery: off` globally. The generated output matches stock
  sqlc-gen-go; no metaquery import, no metadata files.

## CLI reference

```
demo seed                       Populate sample authors + posts.
demo list   [flags]             List authors via mqpgx.Run (untyped Result).
demo list --typed [flags]       List authors via mqpgx.Scan[db.Author] (typed).
demo counts [flags]             Post counts per author (aggregation via Scan[T]).
```

Common flags:

| Flag        | Meaning                                                               |
| ----------- | --------------------------------------------------------------------- |
| `--db`      | `DATABASE_URL`. Default `postgres://demo:demo@localhost:5544/demo?sslmode=disable` |
| `--where`   | JSON array of `metaquery.Filter`                                      |
| `--order`   | JSON array of `metaquery.OrderBy`                                     |
| `--page`    | 0-indexed page                                                        |
| `--size`    | Page size (`0` = unlimited)                                           |
| `--total`   | Also run a `COUNT(*)` to populate `meta.pagination.total` (default `true` when `--size > 0`) |

`--where` payload shape (each entry is a `metaquery.Filter`):

```json
[
  {"column": "name", "op": "ILIKE", "value": "%ada%"},
  {"expr":   "created_at > now() - interval '7 days'"}
]
```

`--order` payload shape:

```json
[
  {"column": "name", "dir": "ASC"}
]
```

## Example output

### `demo list --size 3 --where '[{"column":"name","op":"ILIKE","value":"%a%"}]'`

```jsonc
{
  "data": [
    [1, "Ada Lovelace",      "First programmer.", "2026-04-20T17:50:12.123456Z"],
    [2, "Alan Turing",       "Father of CS.",     "2026-04-20T17:50:12.234567Z"],
    [3, "Barbara Liskov",    null,                "2026-04-20T17:50:12.345678Z"]
  ],
  "meta": {
    "columns": [
      {"name":"id",         "go_type":"int64",              "db_type":"bigserial",  "not_null":true, "table":"authors"},
      {"name":"name",       "go_type":"string",             "db_type":"text",       "not_null":true, "table":"authors"},
      {"name":"bio",        "go_type":"pgtype.Text",        "db_type":"text",                        "table":"authors"},
      {"name":"created_at", "go_type":"pgtype.Timestamptz", "db_type":"timestamptz","not_null":true, "table":"authors"}
    ],
    "pagination": {"limit": 3, "total": 4},
    "where":   [{"column":"name","op":"ILIKE","value":"%a%"}],
    "order_by":[{"column":"name","dir":"ASC"}]
  }
}
```

### `demo counts`

```jsonc
{
  "data": [
    {"author_id": 1, "total": 3},
    {"author_id": 2, "total": 5},
    {"author_id": 3, "total": 2},
    {"author_id": 4, "total": 4},
    {"author_id": 5, "total": 1}
  ],
  "meta": {
    "columns": [
      {"name":"author_id", "go_type":"int64", "db_type":"int8", "not_null":true, "table":"posts"},
      {"name":"total",     "go_type":"int64", "db_type":"bigint"}
    ],
    "group_by": ["author_id"]
  }
}
```

## What the demo demonstrates

- **Metadata emission** — `db/query.sql.metaquery.go` exposes `MetaListAuthors`,
  `MetaListPosts`, etc., each carrying the SQL, columns, args.
- **Untyped `Run`** — `list` takes arbitrary user filters parsed from JSON
  and projects the full row shape as `[][]any` with a self-describing
  `Meta.Columns` — ready to hand to a generic API client / admin UI.
- **Typed `Scan[T]` (model reuse)** — `list --typed` scans directly into
  the sqlc-generated `db.Author` model.
- **Typed `Scan[T]` (aggregation)** — `counts` groups by `author_id`,
  projects `count(*) AS total`, and scans into a user-defined
  `AuthorPostCount` struct. Shape is validated at `Build` time; drift
  errors before any query runs.
- **Bidirectional filter metadata** — `meta.where` / `meta.order_by` /
  `meta.pagination` round-trip the user's request back in the response
  envelope.

## Troubleshooting

- `connection refused` on port 5544 → `make up` first (pg listens on the
  host's `5544`, not `5432`, to avoid colliding with a local postgres).
- `plugin binary not found` → `make plugin`, or `make -C ../..
  bin/sqlc-go-codegen-metaquery`.
- Schema changed, need to reseed → `make down && make demo`.
