# Issue 0001: dataset.Response[T] embeds internal Meta (over-exposes internals)

Found while adopting `metaquery/dataset` in redline2 (a Huma+gat API where the
Response is serialized directly to clients).

## What
`dataset.Response[T]` is:

```go
type Response[T any] struct {
    Data     []T            `json:"data"`
    Count    *Count         `json:"count,omitempty"`
    Meta     metaquery.Meta `json:"meta"`
    Rendered Rendered       `json:"rendered"`
}
```

`metaquery.Meta` carries `Columns`, `Pagination`, **`Where []Filter`**, `Having`,
`GroupBy`, `OrderBy`. `Where` includes the *compiled* search predicates — raw
`Expr` strings, column names, and the **bound argument values** the user
searched for.

## Why it's a problem
For an API that serializes `Response` straight to the wire (the common case for
"DataSet" endpoints), this:
- leaks query internals (raw SQL exprs, internal column names) into the public
  response, and
- echoes user input back inside `Meta.Where` (redundant with `Rendered`, and
  unexpected in a list payload).

redline2 had to define its own slim response and map
`{Data, Count, Rendered.Search}` past `Meta` to avoid shipping it.

## Suggestion
One of:
1. Drop `Meta` from `Response` (keep `Data`, `Count`, `Rendered`); expose Meta
   only via `Shape`/`Builder.Meta()` for callers who want it.
2. Split: `Response` stays lean; add `ResponseWithMeta[T]` (or a `WantMeta`
   flag) for introspection use cases.
3. At minimum, make `Meta` `omitempty` / opt-in and document that it exposes
   compiled filters incl. arg values.

## Severity
Low (workaround is a 3-field mapping) but it's a footgun for the headline
use case (serialize the DataSet response directly).

## Resolution (done)
Took suggestion 1: dropped `Meta` from `dataset.Response[T]` — it now carries
only `Data`, `Count`, `Rendered`. Callers needing column/filter introspection
use `Builder.Meta()` or the scan `TypedResult`. Doc on `Response` notes why Meta
is intentionally absent.
