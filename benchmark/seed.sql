-- 1M-row table for planner comparisons. category has 100 distinct values
-- (~10k rows each, 1% selectivity); views/created_at are indexed for the
-- ordered cases.
DROP TABLE IF EXISTS bench;
CREATE TABLE bench (
  id         bigserial PRIMARY KEY,
  name       text        NOT NULL,
  category   int         NOT NULL,
  views      bigint      NOT NULL,
  created_at timestamptz NOT NULL
);
INSERT INTO bench (name, category, views, created_at)
SELECT 'name_' || g,
       (g % 100),
       (g * 2654435761 % 100000),
       now() - ((g % 1000000) || ' seconds')::interval
FROM generate_series(1, 1000000) g;
CREATE INDEX bench_category_idx ON bench(category);
CREATE INDEX bench_views_idx    ON bench(views);
CREATE INDEX bench_created_idx  ON bench(created_at);
ANALYZE bench;
SELECT count(*) AS rows FROM bench;
