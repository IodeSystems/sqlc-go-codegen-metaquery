# sqlc-go-codegen-metaquery

A fork of [`sqlc-gen-go`](https://github.com/sqlc-dev/sqlc-gen-go) that emits
per-query **metadata** alongside the usual sqlc output, plus a small runtime
that wraps any query as a CTE so you can compose filters, ordering,
pagination, and aggregations dynamically — without ever modifying the
original SQL.

Built for **Go + Postgres + pgx/v5 + sqlc**. Drop-in compatible with stock
sqlc-gen-go: the existing `q.ListAuthors(ctx)` methods are unchanged; the
metaquery surface is purely additive.

```go
// Your sqlc query:
// -- name: ListAuthors :many
// SELECT id, name, bio FROM authors ORDER BY name;

// Wrap it, add filters/pagination, run it — all at runtime:
res, _ := mqpgx.Scan[db.Author](ctx, conn,
    db.WrapListAuthors().
        ApplyFilter(db.ListAuthorsCols.Name.ILike("%ada%")).
        ApplyOrder(db.ListAuthorsCols.CreatedAt.Desc()).
        ApplyPagination(metaquery.PageRequest{Page: 0, Size: 20, Total: true}))

// res.Data is []db.Author
// res.Meta has the applied Filter/OrderBy/Pagination + column metadata,
// ready to serialize as an HTTP response body.
```

See [`examples/pgx/`](examples/pgx/) for a full runnable demo (docker-compose
pg17, migrations, seed, JSON-output CLI).

## Why use this

Stock sqlc is great for static queries but painful once you need dynamic
filtering, pagination, or admin UIs on top. Typical workarounds:
- Write N variants of a query — explodes quickly
- Drop to `database/sql` + a query builder — loses sqlc's type safety
- Build your own metadata indirection — what this project is

This fork keeps sqlc's compile-time query validation and generated types,
adds a typed wrapper per query, and gives you a runtime API that's
JSON-round-trippable for HTTP endpoints and admin consoles.

## Installation

### Via WASM (recommended)

<!-- release:begin -->

### Latest release — v0.1.0

```yaml
plugins:
- name: metaquery
  wasm:
    url: https://github.com/IodeSystems/sqlc-go-codegen-metaquery/releases/download/v0.1.0/sqlc-go-codegen-metaquery.wasm
    sha256: f8d79989e58225905fd42adaa2b55e055770ac0f9e46bcdfe028fafbd865a805
```

<!-- release:end -->

### From source

```sh
git clone https://github.com/iodesystems/sqlc-go-codegen-metaquery.git
cd sqlc-go-codegen-metaquery
make all       # produces bin/sqlc-go-codegen-metaquery (+ .wasm)
```

Point sqlc at the binary:

```yaml
# sqlc.yaml
version: '2'
plugins:
- name: metaquery
  process:
    cmd: /path/to/sqlc-go-codegen-metaquery
sql:
- schema: schema.sql
  queries: query.sql
  engine: postgresql
  codegen:
  - plugin: metaquery
    out: db
    options:
      package: db
      sql_package: pgx/v5
      emit_db_tags: true
      emit_json_tags: true
      emit_metaquery: cols    # off | meta | wrap | cols (default: cols)
```

Then `sqlc generate` produces the usual `db/query.sql.go` + `db/models.go` +
a new `db/query.sql.metaquery.go` carrying the per-query metadata and typed
helpers.

In your Go code, import the runtime:

```go
import (
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqpgx"
)
```

The `mqpgx` adapter lives in a sibling Go module so users of the core
`metaquery` package with a different driver aren't forced to pull in pgx.

## What it emits

For each query sqlc processes, three symbols (at the default `cols` level):

| Symbol | Purpose |
| --- | --- |
| `MetaListAuthors metaquery.Query` | Runtime-readable metadata: the SQL text, columns (Name, GoType, DBType, NotNull, Table, …), args, source file, etc. JSON-serializable. |
| `WrapListAuthors(args...) *metaquery.Builder` | Typed constructor that binds the original query's positional args at compile time and returns a Builder ready for filter/order/pagination/aggregation methods. |
| `ListAuthorsCols struct{...}` | Typed column references. `ListAuthorsCols.Name` is a `TextCol`; `ListAuthorsCols.ID` is an `IntCol`. Each exposes ops appropriate to its type (`ILike(string)` on text, `Gt(int64)`, `Between(int64, int64)` on int, etc.). Column names and op/value types are compile-time checked. |

Six column kinds ship by default (Text/Int/Float/Bool/Time/Bytes), with
`AnyCol` as the escape hatch for arrays, enums, pgtype.Numeric, UUIDs, etc.

## Emission levels

You pay only for what you use. Set globally via `emit_metaquery` or per-query
with `-- metaquery: <level>`:

| Level | Emits | Typical use |
| --- | --- | --- |
| `off` | nothing | Query is only called via the regular sqlc method; no dynamic wrapping needed |
| `meta` | `Meta<Name>` only | You want runtime introspection (schema export, generic API handler) without the builder surface |
| `wrap` | `+ Wrap<Name>(args...)` | Typed wrappers but filter against column names as strings |
| `cols` *(default)* | `+ <Name>Cols` | Full Tier 2 — typed wrappers + typed column refs |

Per-query override example:

```sql
-- name: TruncateAll :exec
-- metaquery: off
-- (Only called from seed/migration code; no builder needed.)
TRUNCATE authors RESTART IDENTITY CASCADE;
```

## Safety surface

| Failure mode | Caught | How |
| --- | --- | --- |
| Wrong wrapper arg type | compile time | `WrapGetAuthor("not an int")` won't compile |
| Unknown column in `.Where`/`.OrderBy`/etc | compile time (via typed cols) or pre-query (via whitelist) | `ListAuthorsCols.Typo` → compile error; `.Where("typo",...)` → Build-time error |
| Wrong op for column type | compile time (typed cols) or pre-query (ValidateFilter) | `IntCol` has no `.ILike`; `Filter{Op: OpILike}` on an int column → `op "ILIKE" not valid for column "id" (int64/int)` |
| Wrong value type | compile time (typed cols) or pre-query (ValidateFilter) | `IntCol.Eq` takes `int64`; JSON-driven filters with string values for int columns are rejected before any query runs |
| Scan-struct shape drift | pre-query | `Validate[T]` reflects on T and diffs against `b.OutputColumns()` |
| Malformed raw SQL in `WhereExpr`/`Agg` | query time | Passed through to Postgres verbatim; caller owns safety |

The validator is also callable as a library function (`metaquery.ValidateFilter(q, f)`)
so JSON-driven HTTP handlers can fail fast on bad client input, using the
same rules the builder applies internally.

## Relationship to upstream

This is a clean fork of
[`github.com/sqlc-dev/sqlc-gen-go`](https://github.com/sqlc-dev/sqlc-gen-go),
not a registered GitHub fork. Upstream is tracked as a git remote:

```sh
git remote -v
# origin    git@github.com:IodeSystems/sqlc-go-codegen-metaquery.git
# upstream  https://github.com/sqlc-dev/sqlc-gen-go.git
```

To sync upstream updates:

```sh
git fetch upstream
git merge upstream/main   # resolve conflicts in the three touched files
                          #   (internal/gen.go, internal/result.go, go.mod + Makefile)
                          # import-path conflicts resolve mechanically with sed.
```

## Caveats / non-goals

- **Postgres + pgx/v5 only.** MySQL and SQLite aren't blocked by the runtime
  design but aren't implemented; the `mqpgx` adapter is pgx-specific.
- **The builder wraps, never rewrites.** A filter references the *output*
  columns of the original query. If you need to filter on a column the
  query doesn't project, either widen the query or use `WhereExpr`.
- **`Sum`/`Avg`/`Min`/`Max` aggregates project the source column's type
  without null-safety** — aggregates over empty groups return NULL. Use
  `Agg("x", "coalesce(sum(y), 0)", "int64")` as the escape hatch.

## License

MIT — same as upstream. See [LICENSE](LICENSE).
