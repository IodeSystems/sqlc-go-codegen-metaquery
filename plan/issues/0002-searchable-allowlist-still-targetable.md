# Issue 0002: Config.Searchable gates global terms but not field:-targeting

Found while adopting `metaquery/dataset` in redline2; affects fidelity with
redline's DataSet ("a column with no search config is not searchable at all").

## What
`Config.Searchable` (allowlist) restricts which columns **unqualified** terms
hit. But a column NOT in `Searchable` is still reachable via `field:value`
targeting — confirmed by the package's own test
`TestShape_SearchableStillTargetable` (`bio` excluded from free-text, yet
`bio:hello` still matches).

## Why it matters
redline's model: a field with no `search { … }` config is **not searchable by
any path** (neither global nor targeted). With the current allowlist, porting
"only NAME is searchable" still leaves `description:foo`, `bio:foo`, etc.
targetable — a behavior change vs the source app, and a mild
information-disclosure surface (probe arbitrary text columns by target).

To get strict parity today, the caller must additionally enumerate every
non-allowlisted column and set `search.Config.Fields[col] = {Disable: true}` —
duplicating what `Searchable` already implies.

## Suggestion
Either:
1. Add a strict mode (e.g. `Config.SearchableExclusive bool`, or treat a
   non-empty `Searchable` as a hard allowlist) where non-listed columns are
   neither global nor targetable; or
2. Auto-`Disable` columns absent from a non-empty `Searchable` (so `Searchable`
   means "these and only these are searchable, both global and targeted"); or
3. If the current behavior is intentional (targeting-any is a feature), document
   it prominently and provide a one-liner helper to build the strict allowlist.

## Severity
Low–medium: it's a faithful-port and surface-area concern, not a crash. Current
workaround is enumerating `Fields{Disable:true}` from the builder's columns.

## Resolution (done)
Took suggestion 2 (hard allowlist) plus the redline `global=false` case:

- A non-empty `Config.Searchable` (or new `Config.Targetable`) is now a HARD
  allowlist. Columns in `Searchable` are global + targetable; columns in
  `Targetable` are targeted-only; every other column is `Disable`d (not
  searchable by any path). Explicit `search.Config.Fields` entries still win.
- `bio:hello` with `Searchable:["name"]` no longer touches bio — the unknown
  target falls back to a global value search on the allowed columns.
- Tests: `TestShape_SearchableIsHardAllowlist`, `TestShape_Targetable`.
