# sqlc-go-codegen-metaquery

A fork of [`sqlc-gen-go`](https://github.com/sqlc-dev/sqlc-gen-go) that emits
per-query **metadata** alongside the usual sqlc output, plus a small runtime
that wraps any query as a CTE so you can compose filters, ordering,
pagination, and aggregations dynamically — without ever modifying the
original SQL.

Built for **Go + sqlc**, with adapters for **Postgres + pgx/v5** and
**SQLite + database/sql**. Drop-in compatible with stock sqlc-gen-go: the
existing `q.ListAuthors(ctx)` methods are unchanged; the metaquery surface
is purely additive.

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
pg17, migrations, seed, JSON-output CLI), or [`examples/sqlite/`](examples/sqlite/)
for the same demo against a local SQLite file (pure-Go driver, no Docker).

## Why this exists

Stock sqlc is excellent for static queries and falls over the moment a screen
needs dynamic filtering, sorting, pagination, or faceting — exactly what every
list view and admin console needs. The query text is fixed at codegen time, so
"filter by whichever columns the user picked, in whatever order, with
operators chosen at runtime" has no first-class answer.

The existing ways out each give something up:

| Approach | What you give up |
| --- | --- |
| **Hand-write N query variants** | Combinatorial explosion; every new filter/sort multiplies the query count. Unmaintainable past a few dimensions. |
| **`col = @x OR @x IS NULL` / `CASE` tricks** | The planner can't use indexes well, and you still can't express "operator chosen at runtime" or dynamic `ORDER BY`. |
| **sqlc macros (`sqlc.slice`, `narg`)** | Covers `IN`-lists and nullable args only. There is **no** `sqlc.switch`/dynamic-`WHERE` macro in shipped sqlc — it's a long-standing proposal, and even as proposed it generates a *fixed set of compile-time variants*, not arbitrary runtime composition. |
| **Drop to an ORM / query builder** (GORM, Bun, Ent, SQLBoiler, go-jet, Squirrel) | You leave the sqlc model entirely: hand-written SQL stops being the source of truth, and you trade sqlc's "the SQL is real and validated against the schema" guarantee for the ORM's abstraction (and its query-shape footguns). |
| **Runtime "dynamic sqlc" libs** (e.g. [`getangry/sqld`](https://pkg.go.dev/github.com/getangry/sqld)) | Conditions are injected into your SQL at hand-placed annotation comments; fields are string-keyed with types inferred at runtime/from naming; rows are scanned by reflection. Keeps sqlc, loses compile-time column/op safety. |

This project takes a different path: **a codegen plugin (fork of
sqlc-gen-go) that emits, per query, typed column references plus a runtime
builder that wraps the original query as a CTE.** You keep sqlc's validated SQL
and generated row types, and you get a composition surface that is checked by
the Go compiler.

```go
// Wrong op or value type for the column is a COMPILE error, not a runtime one:
db.ListAuthorsCols.Name.ILike(5)      // ✗ ILike wants string
db.ListAuthorsCols.ID.ILike("%x%")    // ✗ IntCol has no ILike
db.ListAuthorsCols.Name.ILike("%ada%")// ✓
```

### How it compares

| | this project | runtime libs (sqld) | ORMs / builders | sqlc macros / `sqlc.switch` (proposed) |
| --- | --- | --- | --- | --- |
| **Keeps sqlc's SQL + types** | ✅ | ✅ | ❌ replaces them | ✅ |
| **Compile-time column/op/value safety** | ✅ typed cols | ❌ string keys | ⚠️ builder-typed, schema-derived | ⚠️ only the fixed variants |
| **Runtime-arbitrary filter/order/page** | ✅ | ✅ | ✅ | ❌ fixed variants only |
| **Works on any query unmodified** | ✅ CTE wrap, original SQL untouched | ❌ annotate each query | n/a | ❌ rewrite the query |
| **JSON-round-trippable (HTTP in, response out)** | ✅ | ⚠️ partial | ❌ | ❌ |
| **Setup cost** | codegen plugin + upstream fork to track | drop-in runtime dep | new dependency + paradigm | built into sqlc |

The honest tradeoff: this is a **fork of sqlc-gen-go**, so you run it through
`sqlc generate` and track upstream (see
[Relationship to upstream](#relationship-to-upstream)). Runtime-only libraries
have no fork to maintain. Choose this when compile-time column/op safety and
zero-per-query-setup wrapping are worth the codegen step — which is the whole
point for admin UIs and list endpoints, where the filterable surface is wide
and string-keyed mistakes are otherwise caught only in production.

See [`benchmark/`](benchmark/) for measured proof the CTE wrap is free on
modern Postgres ([Performance](#performance)), and
[DataSet endpoints](#dataset-batteries-included-search-endpoints) for the
downstream wrapper that turns one wrapped query into a complete
search/sort/paginate/facet endpoint with a Google-style search DSL.

## Installation

### Via WASM (recommended)

<!-- release:begin -->

### Latest release — v1.1.1

```yaml
plugins:
- name: metaquery
  wasm:
    url: https://github.com/IodeSystems/sqlc-go-codegen-metaquery/releases/download/v1.1.1/sqlc-go-codegen-metaquery.wasm
    sha256: feb4d49fa9cc10945685b888790e4467aad8828b7bb43ad5711b99e45e8e7f52
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

In your Go code, import the runtime and pick the adapter for your engine:

```go
// Postgres + pgx/v5:
import (
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqpgx"
)

// SQLite + database/sql:
import (
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery"
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqsqlite"
    _ "modernc.org/sqlite" // or any database/sql sqlite driver
)
```

`mqsqlite` is driver-agnostic — it talks to `database/sql`, so it works with
`modernc.org/sqlite` (pure Go), `mattn/go-sqlite3`, libsql, etc. The wrapped
query's `Dialect` is set automatically by the codegen based on
`engine: sqlite` in `sqlc.yaml`, and the builder emits `?N` placeholders to
match.

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

## DataSet: batteries-included search endpoints

The `metaquery` builder is the primitive. The
[`metaquery/dataset`](metaquery/dataset) package is the downstream wrapper that
turns one wrapped query into a complete list endpoint — search, partition,
multi-column ordering, pagination, per-column capability metadata — from a
single JSON request, in one call:

```go
import (
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/dataset"
    "github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqpgx"
)

// One request shapes the whole query; one response serializes straight to the client.
res, err := dataset.Run(ctx, db.WrapListUsers(orgID), req, cfg,
    func(ctx context.Context, b *metaquery.Builder) (*metaquery.TypedResult[UserView], error) {
        return mqpgx.Scan[UserView](ctx, pool, b)
    })
// res.Data []UserView, res.Count {inQuery,inPartition}, res.Columns [capabilities], res.Rendered
```

`dataset.Request` is the JSON contract — page/pageSize, ordering, a free-text
`search` string, a `partition` string, and `showCounts`/`showColumns` toggles:

```json
{ "page": 0, "pageSize": 25,
  "ordering": [{"field": "name", "order": "ASC"}],
  "search": "ada status:active, !archived",
  "showCounts": true, "showColumns": true }
```

The **search DSL** (`metaquery/search`) compiles that string against the
wrapped query's output columns into a single bound, parameterized `WHERE` —
never string interpolation. Behavior is type-driven from column metadata, so a
zero `Config` already does the sensible thing:

- bare terms (`ada`) → case-insensitive `contains` across text columns
- `field:value` targets a column; operators ride in the value: `score:>=90`,
  `score:10..99`, `name:foo*` (wildcard), enum/uuid/inet columns → exact match
- `!term` negates; `,` separates OR-groups
- `partition` is a second independent search expression, giving you the
  `inPartition` vs `inQuery` counts that faceted UIs need

`Config` layers policy on top of the defaults: `Searchable`/`Targetable`
allowlists (hard — unlisted columns are unreachable by any search path),
`Orderable` to restrict sortable fields, virtual `Named` searches not bound to
a column (`onlyMine`, `daysAgo`), per-field overrides and aliases, and page-size
caps. `Response.Columns` reports each column's `searchable`/`global`/`orderable`
flags and current sort, so a client can render headers and filter controls
without hardcoding the schema.

Net effect: a typed sqlc query becomes an admin-grade, self-describing
list endpoint — the use case that motivated the whole project — without
hand-writing filter/sort/pagination plumbing per screen.

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

## Performance

Wrapping a query as `WITH __q AS (<original>) SELECT … FROM __q WHERE …` sounds
like it should defeat the planner. It doesn't, on **Postgres 12+**: a
non-recursive CTE referenced once and not `MATERIALIZED` is **inlined** — folded
into the outer query — so outer filters push down and indexes are used exactly
as in hand-written SQL. (The "CTEs are optimization fences" rule is pre-12
folklore.)

Measured on **1,000,000 rows, Postgres 17** ([`benchmark/`](benchmark/),
`./run.sh` to reproduce — `EXPLAIN (ANALYZE, BUFFERS)`):

| Case | Inner query (sqlc) | Runtime adds | Raw | Wrapped |
| --- | --- | --- | --- | --- |
| Indexed equality | `SELECT …` | `WHERE category=42 LIMIT 20` | 0.14 ms | **0.10 ms** — identical plan |
| `count(*)` page total | `SELECT …` | `WHERE category=42` | 0.70 ms | **0.67 ms** — identical (index-only scan) |
| Inner `LIMIT 1000` fence | `… ORDER BY views DESC LIMIT 1000` | `WHERE category=42` | 0.50 ms | **0.29 ms** — identical (raw subquery fences the same way) |

The wrapped plan is the same shape as hand-written SQL — to the buffer.

**The one real cost — doubled `ORDER BY`.** If the sqlc query keeps its own
`ORDER BY` *and* you apply ordering at runtime, you emit a redundant nested
sort that can flip the planner off the index:

| Inner `ORDER BY` | Outer `ORDER BY` | Plan | Time |
| --- | --- | --- | --- |
| yes | no | index scan, stop at 20 | 0.26 ms |
| no | yes | index scan, stop at 20 | 0.23 ms |
| **yes** | **yes** | bitmap → materialize 10k → top-N sort | **12.7 ms** |

A plain (non-CTE) subquery with the same doubled sort is just as slow — so this
is **not** the wrap, it's the redundant ordering. **Rule: drop `ORDER BY` from
any query you intend to wrap dynamically, and order via `.ApplyOrder(...)`.**
The runtime order is all you need (rows 1–2 above), and codegen warns when a
wrapped query ends in `ORDER BY`. Inner `LIMIT`/`GROUP BY`/`DISTINCT` remain
genuine optimization fences — but they fence hand-written subqueries identically.

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

- **Postgres (pgx/v5) and SQLite (database/sql) are supported.** MySQL isn't
  blocked by the runtime design but doesn't yet have an adapter.
- **`ILIKE` on SQLite** is auto-translated to `LIKE` at Build time. SQLite's
  `LIKE` is case-insensitive for ASCII by default — same effective behavior
  as Postgres `ILIKE` for typical text columns. For Unicode-aware
  case-insensitive matching, use `WhereExpr` with `LOWER(...)`.
- **The builder wraps, never rewrites.** A filter references the *output*
  columns of the original query. If you need to filter on a column the
  query doesn't project, either widen the query or use `WhereExpr`.
- **Don't leave `ORDER BY` in a query you intend to wrap dynamically.** The
  CTE wrap itself is free on Postgres 12+ (it inlines the CTE, so filters push
  down and indexes are used — see [`benchmark/`](benchmark/) for measured
  plans). The one real cost: if the sqlc query keeps its own `ORDER BY` *and*
  you apply ordering at runtime, you emit a doubled nested sort that can flip
  the planner off the index (≈44× slower in the benchmark's 1M-row case). Drop
  `ORDER BY` from the query and always order via `.ApplyOrder(...)`; the
  runtime order is all you need. Inner `LIMIT`/`GROUP BY`/`DISTINCT` remain
  genuine optimization fences — but they fence hand-written subqueries the same
  way.
- **`Sum`/`Avg`/`Min`/`Max` aggregates project the source column's type
  without null-safety** — aggregates over empty groups return NULL. Use
  `Agg("x", "coalesce(sum(y), 0)", "int64")` as the escape hatch.

## License

MIT — same as upstream. See [LICENSE](LICENSE).
