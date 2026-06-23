-- Each case: RAW (hand-written) vs WRAPPED (metaquery CTE form).
-- Inner query mirrors what sqlc emits; wrapper adds WHERE/ORDER/LIMIT on __q.

\echo ====================================================================
\echo CASE A — indexed equality, inner query has NO order/limit
\echo --- RAW
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
SELECT id, name, category, views, created_at FROM bench WHERE category = 42 LIMIT 20;
\echo --- WRAPPED
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
WITH __q AS (SELECT id, name, category, views, created_at FROM bench)
SELECT * FROM __q WHERE category = 42 LIMIT 20;

\echo ====================================================================
\echo CASE B — REALISTIC: inner sqlc query HAS "ORDER BY created_at DESC"
\echo --- RAW (what you would hand-write)
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
SELECT id, name, category, views, created_at FROM bench
  WHERE category = 42 ORDER BY created_at DESC LIMIT 20;
\echo --- WRAPPED
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
WITH __q AS (SELECT id, name, category, views, created_at FROM bench ORDER BY created_at DESC)
SELECT * FROM __q WHERE category = 42 ORDER BY created_at DESC LIMIT 20;

\echo ====================================================================
\echo CASE C — FENCE: inner sqlc query already has its own LIMIT (top-1000)
\echo --- RAW equivalent (filter the top-1000 by views)
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
SELECT * FROM (SELECT id, name, category, views, created_at FROM bench ORDER BY views DESC LIMIT 1000) t
  WHERE category = 42;
\echo --- WRAPPED
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
WITH __q AS (SELECT id, name, category, views, created_at FROM bench ORDER BY views DESC LIMIT 1000)
SELECT * FROM __q WHERE category = 42;

\echo ====================================================================
\echo CASE D — pagination COUNT(*): raw vs metaquery BuildCount form
\echo --- RAW
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
SELECT count(*) FROM bench WHERE category = 42;
\echo --- WRAPPED (metaquery BuildCount)
EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF)
WITH __q AS (SELECT id, name, category, views, created_at FROM bench)
SELECT count(*) FROM __q WHERE category = 42;
