#!/usr/bin/env bash
# Reproduce the planner comparison: raw hand-written SQL vs the metaquery
# CTE-wrapped form, across query shapes. Requires docker.
set -euo pipefail
cd "$(dirname "$0")"

PSQL=(docker compose exec -T db psql -U bench -d bench -q)

docker compose up -d
printf "waiting for pg..."
until docker compose exec -T db pg_isready -U bench -d bench >/dev/null 2>&1; do
  printf "."; sleep 1
done
echo " ready"

echo "== seeding 1M rows =="
"${PSQL[@]}" -f - < seed.sql

echo "== cases =="
"${PSQL[@]}" -f - < cases.sql
