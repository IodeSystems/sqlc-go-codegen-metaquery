# Issue 0003: `unnest` — multi-arg is rejected, single-arg silently boxes to `interface{}`

Found while building per-processor reconciliation cutoffs in redline2 (sup #134).

Two related defects, and the second is the more valuable one:

1. **Multi-argument `unnest(a, b)` is rejected outright** — codegen fails.
2. **Single-argument `unnest(x)` succeeds but emits `interface{}`** for the unnested
   column, with no warning. This is the one that will quietly cost someone a debugging
   session.

## 1. Multi-arg is rejected

PostgreSQL allows a multi-argument `unnest` in the FROM clause, where it behaves like
`ROWS FROM (...)` and zips the arrays together — the natural way to pass a parallel
key/value list as two parameters. sqlc rejects it:

```
q/multi.sql:2:15: function unnest(unknown, unknown) does not exist
```

### It is the ARITY, not parameters or casts

A repro with **literal, explicitly cast arrays and no `sqlc.arg` at all** still fails, so
this is not parameter inference and not cast propagation:

```sql
-- name: LiteralMultiArg :many
SELECT * FROM unnest(ARRAY['a','b']::text[], ARRAY[1,2]::int[]) AS u(provider, hour);
```

Full repro: `schema.sql` with any table, the query above, and

```yaml
version: "2"
sql:
  - engine: postgresql
    schema: schema.sql
    queries: q
    gen:
      go:
        package: repro
        out: gen
```

## 2. Single-arg silently boxes the result column

This form generates fine — and is the workaround for #1 — but look at the output:

```sql
-- name: SingleArgUnnest :many
WITH lookup AS (
    SELECT pv.provider, hr.hour
    FROM unnest(sqlc.arg(providers)::text[]) WITH ORDINALITY AS pv(provider, i)
    JOIN unnest(sqlc.arg(hours)::int[])      WITH ORDINALITY AS hr(hour, i) USING (i)
)
SELECT t.id, l.hour FROM t JOIN lookup l ON l.provider = t.provider;
```

```go
type SingleArgUnnestParams struct {
	Providers []string      // args resolve — the cast is on the parameter
	Hours     []int32
}
type SingleArgUnnestRow struct {
	Hour interface{}        // ← anyelement NOT resolved, silently boxed
}
```

The **arguments** type correctly; the **result column** does not. `unnest(anyarray)` returns
`SETOF anyelement` and sqlc does not resolve the element type, so it falls back to
`interface{}` without saying so.

An explicit outer cast rescues it (`COALESCE(l.hour, …)::int` → `int32`), which is the same
rule this codebase already documents for `citext` and for `COALESCE((subquery), '')`. The
problem is that nothing tells you — the box is silent, and it surfaces as a scan failure or
a useless field far from the query.

## Where this lives — read before picking a fix

The rejection in #1 comes from **sqlc core's query analysis**, not from this plugin. Two
checks:

- The repro was run with plain `sqlc generate` and the **built-in** Go codegen
  (`gen: go:`). This plugin was never configured or invoked; the error is identical.
- `go version -m $(which sqlc)` reports `mod github.com/sqlc-dev/sqlc v1.30.0`. redline2's
  `db/sqlc.yaml` wires us in as a `plugins:` process entry, so sqlc analyses the SQL and
  only then hands us the result.

By the time this plugin receives a `GenerateRequest`, the query has already been rejected.
Today we fork `sqlc-gen-go` (codegen), not `sqlc` (core + catalog) — so anything about what
SQL *means* lands outside the fork, and only what Go *comes out* is ours.

### There is no catalog row to copy

Worth knowing before anyone reaches for "just add the signature": PostgreSQL has **no
multi-argument `unnest` in `pg_proc`**. Only three single-arg entries exist:

```
pronargs |         args          |    ret     | proretset
       1 | {anyarray}            | anyelement | t
       1 | {tsvector}            | record     | t
       1 | {anymultirange}       | anyrange   | t
```

The multi-arg form is a **grammar** special case, expanded like `ROWS FROM`, and PG itself
rejects it outside FROM with nearly our error text:

```
ERROR:  function unnest(text[], integer[]) does not exist
```

So adding a catalog entry would not be mirroring Postgres — it would be inventing a row PG
does not have, and it would wrongly make the form legal in `SELECT`. Correct handling
belongs in the FROM-clause path.

### And the columns are necessarily nullable

The multi-arg form **pads short arrays with NULL**:

```
unnest(ARRAY['x','y','z']::text[], ARRAY[1]::int[])
 x | 1
 y |          ← NULL
 z |          ← NULL
```

Nothing can prove the arrays are equal-length, so every column from a multi-arg `unnest` is
nullable. Inferring `NOT NULL` (`string`/`int32`) would be wrong; it has to be
`pgtype.Text`/`pgtype.Int4` (or `*string`/`*int32`).

## Options

**Preferred: warn on an unresolved polymorphic return, and box it `any`.** This is mostly
*adding the warning*, because the boxing already happens — see #2. Extending the same policy
to the multi-arg form turns #1 from a hard failure into a warning plus `any`, and the
existing outer-cast escape hatch recovers a concrete type. Cheapest path that fixes both
defects, and it makes the `citext` / `COALESCE` traps visible at codegen instead of at
runtime. Caveat: the multi-arg rejection happens in core, so the policy has to live where
the rejection does — the warning cannot be added from the plugin.

Alternatives, for the record:

- **Handle the multi-arg form properly in the FROM path** (zip element types, mark all
  columns nullable). Correct and more work; unnecessary if the warn-and-box policy lands.
- **Upstream it to `sqlc-dev/sqlc`.** No fork surface added, but on someone else's
  schedule.
- **Document and move on.** Cheapest, leaves the silent box in #2 unaddressed, which is the
  part most likely to bite.

Whichever is chosen, an error/warning message that names the outer-cast workaround would
save the next person the time this cost.

## Workaround

Two single-argument `unnest`es joined on `WITH ORDINALITY`, plus an **outer cast on the
unnested column** so it does not box (see #2).

redline2 no longer relies on this: migration `000077` replaced the whole approach with a
typed `processor_cutoff` table joined normally, so nothing polymorphic reaches codegen
there. The workaround is recorded because the shape is common, not because it is still in
use.

## Severity

Low for #1 — loud failure at codegen, mechanical workaround, cost is discovery time.

**Medium for #2** — silent `interface{}` in generated code, no warning, and the failure
appears far from its cause. That is the one worth fixing even if #1 is only documented.
