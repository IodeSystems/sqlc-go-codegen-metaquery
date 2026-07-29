# sqlc-go-codegen-metaquery plan

Fork of `sqlc-gen-go` (upstream commit preserved) that additionally emits
per-query metadata so a runtime library can wrap queries as CTEs and compose
filters / aggregations / pagination on top.

## Scope

- Postgres only.
- Original generated query code is untouched. Metadata is additive, emitted to
  a sibling `*.sql.metaquery.go` per source.
- Runtime package (option "b" — generic metadata consumed by a single builder
  lib, no per-query typed builders).

## Architecture

```
          sqlc  ──▶  sqlc-go-codegen-metaquery (plugin)
                            │
                            ├── query.sql.go          (unchanged — original sqlc-gen-go output)
                            ├── models.go             (unchanged)
                            └── query.sql.metaquery.go (NEW — metadata vars)
                                        │
                                        ▼
                        github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery
                        (runtime pkg: Query/Column/Arg types + Builder)
```

## Metadata types (runtime package `metaquery`)

```go
type Query struct {
    Name    string     // Go method name, e.g. "ListAuthors"
    Cmd     string     // ":one", ":many", ":exec", ...
    SQL     string     // Final SQL text sqlc produces (with $1..$N placeholders)
    Source  string     // Source filename
    Columns []Column   // Result columns; empty for :exec
    Args    []Arg      // Positional args in $1..$N order
    Table   *Table     // For :copyfrom; nil otherwise
}

type Column struct {
    Name         string // sqlc-normalized name (Go field name source)
    OriginalName string // Alias or original SQL name
    GoType       string
    DBType       string // e.g. "int8", "text"
    NotNull      bool
    IsArray      bool
    Table        string
    TableAlias   string
}

type Arg struct {
    Position    int    // 1-based
    Name        string // Named param ("@foo") else ""
    GoType      string
    DBType      string
    NotNull     bool
    IsArray     bool
    IsSqlcSlice bool
}

type Table struct{ Catalog, Schema, Name string }
```

## Emission

- New template block `metaqueryFile` in `templates/template.tmpl` rendered via
  `execute(source+".metaquery", "metaqueryFile")` from `gen.go`.
- Output per-source file: `query.sql.metaquery.go` with package-level `var`s:

  ```go
  var MetaListAuthors = metaquery.Query{
      Name: "ListAuthors",
      Cmd:  ":many",
      SQL:  `SELECT id, name, bio FROM authors ORDER BY name`,
      ...
  }
  ```
- Metadata file imports `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery`.

## Runtime builder sketch

```go
b := metaquery.Wrap(&MetaListAuthors /*, baseArgs... */)
b.Where("name", "ILIKE", "%foo%").
    OrderBy("name", "ASC").
    Limit(50).Offset(100)
sql, args, err := b.Build()
// WITH __q AS (SELECT id, name, bio FROM authors ORDER BY name)
// SELECT id, name, bio FROM __q WHERE name ILIKE $1 ORDER BY name ASC LIMIT 50 OFFSET 100
```

Constraints:
- Column names in `.Where`, `.OrderBy`, `.GroupBy`, `.Select` are validated
  against `Q.Columns` (whitelist) — SQL injection defense.
- Original SQL is copied verbatim inside the CTE; its $1..$N placeholders are
  preserved. Added predicate args renumber starting at $N+1. Builder holds
  `base []any` (original) and appends added args in order, concatenating for
  the final `[]any`.
- Aggregation mode: if any `GroupBy` or `Agg` set, projection becomes
  `<groupBy cols>, <aggs>`; else `SELECT <select cols || *> FROM __q`.
- `Having`, `Count()` (wrap in `SELECT count(*) FROM (...) t`), and expression
  filters come later — first pass handles WHERE + ORDER BY + LIMIT/OFFSET.

## Non-goals (first pass)

- MySQL / SQLite support.
- JOINing additional tables into the CTE wrapper (a filter-only wrapper keeps
  the column whitelist sound).
- Rewriting the original SQL.
- Prepared-statement caching of the wrapped CTEs (user responsibility).

## Open questions

- Variable name prefix: `Meta<Method>` vs `<method>Meta` vs `Q<Method>`.
  Going with `Meta<Method>` (reads well, avoids shadowing method names).
- Where should the runtime package live? Same module makes iteration easy but
  forces users to import the plugin repo. For now keep it in this repo under
  `metaquery/`; if awkward, extract later.

## Known issues

Detail lives in `plan/issues/`; these are one-liners so they are discoverable from
here (they were not before).

- [0001](issues/0001-response-embeds-internal-meta.md) — response embeds internal meta.
- [0002](issues/0002-searchable-allowlist-still-targetable.md) — `Searchable` gated
  global terms but not `field:` targeting. **Resolved** (hard allowlist).
- [0003](issues/0003-multi-arg-unnest-rejected.md) — `unnest`: multi-arg is rejected
  outright, and **single-arg silently boxes the result column to `interface{}`**. The
  second is the one worth fixing — the boxing already happens with no warning, so the
  preferred fix (warn on an unresolved polymorphic return, box it `any`) is mostly adding
  the warning. **Two caveats before starting:** the multi-arg rejection is in sqlc CORE,
  before this plugin is invoked, so a warning cannot be added from here; and there is no
  multi-arg `unnest` row in `pg_proc` to copy — it is a grammar special case, so "add the
  signature" would invent an entry PG does not have.
