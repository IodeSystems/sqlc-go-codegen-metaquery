# Planner benchmark — CTE wrap vs hand-written SQL

Does wrapping a query as `WITH __q AS (<original>) SELECT … FROM __q WHERE …`
(what metaquery does) cost anything versus the equivalent hand-written query?

**Short answer:** No — *except* when the wrapped sqlc query carries its own
`ORDER BY` and you also apply ordering at runtime. Then you get a doubled,
nested sort that flips the planner off the index. The CTE itself is free.

## How to run

```sh
./run.sh        # starts pg17, seeds 1M rows, prints EXPLAIN ANALYZE for each case pair
docker compose down -v   # cleanup
```

Setup: 1,000,000 rows; `category` has 100 distinct values (~1% selectivity,
indexed); `views` and `created_at` indexed. Postgres 17, default config.
Plans captured with `EXPLAIN (ANALYZE, BUFFERS, COSTS OFF)`.

## Background — why a CTE wrap is usually free

Pre-12 Postgres treated every CTE as an **optimization fence**: it was
materialized, and the planner could not push predicates from the outer query
into it. Since **PG 12**, a non-recursive CTE that is referenced once and is
not `MATERIALIZED` is **inlined** — folded into the outer query like a
subquery — so quals push through and indexes are used normally.

Inlining still can't push a predicate *below* a construct where that would
change results: an inner `LIMIT`/`OFFSET`, `GROUP BY`/aggregate (becomes
`HAVING`), `DISTINCT`, or window function. Those are genuine fences — but they
fence a hand-written subquery exactly the same way. A plain inner `ORDER BY`
(no limit) is *not* a fence; the planner is free to drop or reorder it.

## Results

| Case | Inner query (sqlc) | Runtime adds | Raw | Wrapped | Verdict |
|---|---|---|---|---|---|
| **A** | `SELECT …` | `WHERE category=42 LIMIT 20` | 0.14 ms / 18 buf | 0.10 ms / 18 buf | **identical** |
| **B** | `… ORDER BY created_at DESC` | `WHERE category=42 ORDER BY created_at DESC LIMIT 20` | 0.29 ms / 33 buf | **12.8 ms / 8348 buf** | **wrapped ~44× slower** |
| **C** | `… ORDER BY views DESC LIMIT 1000` | `WHERE category=42` | 0.50 ms / 1005 buf | 0.29 ms / 1005 buf | identical (both fence on inner LIMIT) |
| **D** | `SELECT …` (count) | `WHERE category=42` | 0.70 ms / 13 buf | 0.67 ms / 13 buf | **identical** (index-only scan) |

A, C, D: the wrapped plan is byte-for-byte the same shape as hand-written.
The CTE wrap costs nothing. C "fences" — but the raw `SELECT … FROM (… LIMIT
1000)` fences identically, because filtering the top-1000 *after* the limit is
the correct semantics, not a metaquery artifact.

### Case B is the one to know about

`ListPosts` in the example is `SELECT … FROM posts ORDER BY created_at DESC` —
sqlc queries very commonly end in `ORDER BY`. Wrap it and apply a runtime
order and you emit:

```sql
WITH __q AS (SELECT … FROM bench ORDER BY created_at DESC)   -- inner sort
SELECT * FROM __q WHERE category = 42 ORDER BY created_at DESC LIMIT 20;  -- outer sort
```

Isolating the cause:

| Variant | Inner ORDER BY | Outer ORDER BY | Plan | Time |
|---|---|---|---|---|
| B1 | yes | no | Index Scan Backward, filter, stop at 20 | 0.26 ms |
| B2 | no | yes | Index Scan Backward, filter, stop at 20 | 0.23 ms |
| **B3** (metaquery form) | **yes** | **yes** | Bitmap on category → materialize 10k → top-N sort | **12.7 ms** |
| B4 (plain subquery, not CTE) | yes | yes | same as B3 | 7.0 ms |

B1 and B2 are fast: either ordering alone lets the planner walk the
`created_at` index and stop after 20 rows. **B3 — both orderings present — is
the slow one**, and B4 shows it has nothing to do with the CTE: a plain nested
subquery with the same doubled `ORDER BY` is just as slow. The redundant inner
sort changes the cost estimate enough that the planner abandons the
index-ordered scan, fetches all 10k category matches, and top-N sorts them.

## Takeaways

1. **The CTE wrap is free on PG 12+.** Equality filters, counts, and
   limit-fenced queries plan identically to hand-written SQL.
2. **Don't put `ORDER BY` in a query you intend to wrap dynamically.** Drop it
   from the sqlc query and always apply ordering through the builder
   (`.ApplyOrder(...)`). The runtime order is all you need, and you avoid the
   doubled-sort penalty entirely (that's case B2: fast).
3. Inner `LIMIT`/`GROUP BY`/`DISTINCT` are real fences, but they fence
   hand-written SQL the same way — not a reason to avoid the wrap.

Recommendation for metaquery itself: detect a trailing inner `ORDER BY` on a
wrapped query and warn at codegen time (or document the rule prominently), so
users don't hit case B silently.
