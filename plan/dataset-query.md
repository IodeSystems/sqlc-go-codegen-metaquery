# Plan: DataSet search language in metaquery

Status: **MVP shipped (phase 1).** Bring redline's DataSet search DSL into the
metaquery runtime, compiled against the per-query **Column metadata** metaquery
already emits. The downstream consumer (redline2's `internal/dataset`) defers its
own search work and adopts this once it lands.

## Built (phase 1)

`metaquery/search/` — ANTLR Go parser (vendored, generated from the shared
`grammar/*.g4`) + AST + compiler + `Apply`:

- `Parse` / error-recovery auto-escape (`searchRendered`) — faithful port of
  dataset's `SearchParser` + `Listener`.
- `compile` — port of `SearchConditionFactory`: AND/OR/NOT/grouping, global vs
  targeted, type-driven defaults (text→ILIKE-contains, int/float/bool→eq+parse,
  unparseable→drop), folded into one `metaquery.Filter{Expr,Args}`.
- Custom fields via **closure + `Col` helper** (`Fn func(Col,string)(*Filter,error)`,
  `return nil,nil` = skip); virtual `Named` searches; `Aliases`; `Scope`;
  `Disable`. Dialect-aware (PG ILIKE / SQLite LIKE), LIKE-metachar escaping.
- `search.Apply(b, query, cfg)` (free function, keeps the antlr dep out of core
  metaquery). `metaquery.Builder.Dialect()` + `metaquery.ValueKind` exposed.
- `make antlr` (`scripts/antlr.sh`) regenerates from grammar; `.g4` kept
  byte-identical to `../dataset` (drift-checkable). Golden tests in
  `search_test.go`.

Deviation from dataset, by design: value-level negation (`field:!val`) is
honored in the **targeted** branch too (dataset honors it only for global
terms) — least-surprise. Unknown `field:` target falls back to a global value
search (matches dataset's runtime, not the plan's earlier "literal" note).

`metaquery/dataset/` — one-call DataSet pipeline over any select:

- `Request` (search + partition + ordering + page) → `Shape(b, req, cfg)` applies
  partition then search (both AND into WHERE), whitelisted ordering, clamped
  pagination. Adapter-agnostic (no DB). `cfg` embeds `search.Config`.
- `Run[T](ctx, b, req, cfg, scan)` shapes + scans via a `ScanFunc` closure (wrap
  `mqpgx.Scan[T]`/`mqsqlite.Scan[T]`), returns `Response[T]{Data,Count,Meta,Rendered}`.
- `RunWithPartitionCount[T]` adds the phase-4 two-pass: a partition-only builder
  is counted for `count.inPartition` (excludes search); base `Run` mirrors
  inPartition==inQuery until then.
- `Config.Searchable []string` mirrors redline's `FieldConfig.Searchable`: when
  set, only those columns are global free-text; the rest are targeted-only
  (explicit `Fields` entries win). The bridge that makes the redline2 swap a
  drop-in.
- `mqpgx.Count` / `mqsqlite.Count` expose a standalone `BuildCount` pass feeding
  `RunWithPartitionCount`.
- This is the runtime helper redline2's `internal/dataset` placeholder adopts;
  optional future codegen can emit a per-query `DataSet<Method>(req)` wrapper
  around it. No sqlc change needed — the plugin already emits the metadata.

Free-text semantics change on adoption: the placeholder did one
`ILIKE '%<whole string>%'`; the DSL tokenizes (`john smith` → AND of two ILIKE,
`,` → OR, `field:val` → targeted), matching redline's original Kotlin behavior.

Remaining: phases 2–4 below.

## Why this belongs in metaquery

redline's Kotlin `DataSet` declares search per field with hand-written lambdas:

```kotlin
field(USER.FIRST_NAME) { search { f, s -> f.containsIgnoreCase(s) } }
field(USER.USER_ID)    { search { f, s -> f.eq(s.toLongOrNull()) } }
field(USER.IS_ADMIN)   { search(global = false) { f, s -> f.eq(s.toBooleanStrictOrNull()) } }
```

metaquery **already knows** each column's `Name`, `GoType`, `DBType`, `NotNull`
(`metaquery.Column`) and has a type-kind→allowed-ops table (`validate.go`
`allowedOps`: text/int/float/bool/time). So those lambdas collapse into
**type-driven defaults**: text→ILIKE-contains, int→eq(parse), bool→eq(parse),
etc. The developer only overrides exceptions. Every sqlc+metaquery project then
gets DataSet-grade search for free — this is the natural home, not each app.

## The grammar (source of truth: `../dataset/src/main/antlr4/*.g4`)

The lexer/parser are ~50 lines of ANTLR4 total. Semantics:

| Syntax              | Meaning                                            |
|---------------------|----------------------------------------------------|
| `john`              | unqualified term — OR across all *global* columns  |
| `john smith`        | space = **AND**                                     |
| `john, jane`        | comma = **OR**                                      |
| `!inactive`         | `!` prefix = **negate** (term or value)            |
| `status:active`     | `field:value` — target one column                  |
| `name:"John Doe"`   | quoted value (`"`/`'`) preserves spaces/specials   |
| `(a, b) status:x`   | parentheses group; precedence                       |
| `field:(a b c)`     | grouped values, implicit AND within                |
| `field:(a, b)`      | grouped values, OR within                          |
| `\!`, `\:`, `\(` …  | backslash escapes a special char                   |

Tokens: `ANY`, `STRING`, `NEGATE !`, `TARGET_SEPARATOR :`, `TERM_OR ,`,
`( )`, `WS`, `ESCAPE \`. **No comparison operators in the grammar** (`<`,`>`,`=`,
`~` are not lexed) — all comparison/range logic lives in per-field value handling
(e.g. redline's `daysAgo:<30` parses `<30` *inside* the field handler). Keep that
property: the grammar stays tiny; operators are a value-coercion concern.

### Parser strategy — DECIDED: Option A (ANTLR Go, shared `.g4`)

**Decision: ANTLR4 Go target generated from the same `DataSetSearch*.g4` that
lives in `../dataset/src/main/antlr4/`.** One grammar, two language targets →
guaranteed fidelity with the Kotlin parser, including ANTLR's error recovery.
The grammar is the single source of truth; metaquery consumes a copy/subtree of
it, never a reimplementation.

Toolchain (matches dataset):
- ANTLR **4.13.2** (dataset pins `antlr4 = "4.13.2"` in `gradle/libs.versions.toml`).
  Go generated code must be produced by the *same major.minor* ANTLR tool — the
  Go runtime version must match the generator version.
- Runtime dep: `github.com/antlr4-go/antlr/v4` (the v4 Go runtime module).
- Generation: `antlr -Dlanguage=Go -package search DataSetSearchLexer.g4
  DataSetSearchParser.g4` into `metaquery/search/`. **Vendor/commit the generated
  parser** so downstream consumers need only the runtime module, not the ANTLR
  tool or a JVM. Wire generation into the `Makefile` (mirror dataset's gradle
  `antlr` task) and mark the output files generated (header) so they're never
  hand-edited.
- Grammar sync: copy the two `.g4` files into `metaquery/search/grammar/` with a
  source-of-truth note pointing at `../dataset`; a Make target regenerates from
  them. (git-subtree is an option but a tracked copy + regen target is simpler
  and matches the repo's lean style.)

Fidelity guard regardless of approach: a **shared golden-test corpus** ported
from dataset's `DataSetBuilderTest` (input → normalized `searchRendered` +
rendered SQL) catches any drift between the two targets.

## AST (mirror dataset's model)

```go
type Conjunction int // And, Or
type Term struct {
    Target    string       // "" = unqualified (global) term
    Negated   bool
    Conj      Conjunction  // relative to the next term
    Values    []TermValue
}
type TermValue struct {
    Value   string
    Negated bool
    Conj    Conjunction  // within the term's value group
}
```

## Compilation: AST × Column metadata → builder conditions

The compiler turns the AST into a tree of metaquery conditions, driven by metadata:

1. **Target resolution.** `field:` → match a `Column` by `Name` (and an optional
   alias map). Unknown target → per config: error, or treat the whole `field:val`
   as a literal global term. Columns are **whitelisted against `query.Columns`**
   → no injection via field names.
2. **Type-kind → default predicate** (from `Column.GoType`/`DBType` via the
   existing `kindOf`):
   - `text`  → `col ILIKE '%' || $ || '%'` (contains, case-insensitive)
   - `int`   → `col = $` (parse to int; unparseable → drop term or no-match)
   - `float` → `col = $`
   - `bool`  → `col = $` (parse true/false)
   - `time`  → eq/range (phase 2)
   - enum (custom `DBType`, `GoType` string) → `col = $::DBType`
3. **Global vs targeted.** Default: text columns are *global* (participate in
   unqualified terms); ids/bools/enums *targeted-only* (require `field:`).
   Overridable per column. An unqualified term expands to `OR` over all global
   columns' default predicates.
4. **Boolean structure.** comma→`OR`, space→`AND`, `!`→`NOT`, parens→nested
   groups. Compile to a `Filter` tree; emit via the builder using parameterized
   `WhereExpr` (values are **always** bound params, never interpolated).
5. **Value operators (phase 2, opt-in per column).** A column may declare it
   accepts comparison/range *values*: `<30`, `>=90`, `10..99`, `*foo` (prefix),
   plus named virtual searches (redline's `daysAgo`, `onlyActive`). These are
   coercion functions `func(raw string) (Filter, ok)` registered per target.

Output is a single composite `Filter`/expression handed to the existing
`ApplyFilter`/`WhereExpr` machinery — no new SQL builder needed.

## Public API (sketch)

```go
// In a new subpackage metaquery/search (keeps the parser/deps isolated).
type ColumnRule struct {
    Searchable bool            // default true
    Global     bool           // participates in unqualified terms
    Op         metaquery.Op   // override default op for this column
    Coerce     func(raw string) (any, bool) // value → bound arg (e.g. parse int)
    Alias      []string       // extra names the `field:` target may use
}
type Config struct {
    Columns map[string]ColumnRule       // by column Name; unset → metadata default
    Named   map[string]func(string)(metaquery.Filter, bool) // virtual searches
}

// Compile parses `query` against b's Column metadata + cfg, returns the
// normalized (auto-escaped) string for DataSetResponse.searchRendered.
func (b *metaquery.Builder) ApplySearch(query string, cfg Config) (rendered string, err error)
```

Defaults come entirely from `b.Meta().Columns` — `Config{}` (zero value) yields
sensible behavior (text columns contains-searchable & global).

## searchRendered + error recovery

Match redline: on a parse error, insert `\` before the offending position and
retry (bounded, e.g. ≤1000). Return the corrected string as `rendered`
(→ `DataSetResponse.searchRendered`) so the UI can show what was actually run.

## partition vs search, and counts

- `partition` and `search` use the **same** compiler; apply partition first, then
  search (two `ApplySearch` calls, or `ApplyPartition`). Both AND into the WHERE.
- `count.inPartition` = rows after partition only; `count.inQuery` = after
  partition+search. metaquery's `WithTotal()` gives one count; supporting both
  needs a partition-only count pass (build the partition-only query and
  `BuildCount`). Add a helper or document the two-pass approach.

## Phasing

1. ~~**MVP**: parser + targeting + text-contains + int/bool eq + AND/OR/NOT/group
   + searchRendered.~~ **DONE.** (alias targets + named virtual searches landed
   here too, ahead of schedule.)
2. **Typed ops**: enum eq (`col = $::DBType`), per-column op overrides beyond the
   defaults. (Per-column `Search` override + aliases already exist; remaining is
   enum/`::DBType` defaulting.)
3. ~~**Value operators**: comparison/range (`<`,`>`,`..`), prefix/wildcard.~~
   **DONE.** `operators.go`: `ParseValueOp` (primitive), `CompareFn` (comparison
   + range), `TimeFn` (date columns, unlocks the time kind), `WildcardFn` (glob
   text + `=` exact), `Col.Between`. Numeric defaults are now operator-aware
   (`score:>=90`, `score:10..99`) — strictly additive (bare value still `=`).
   Validated end-to-end on SQLite. Note for redline: int/time columns must be in
   `Targetable` (hard allowlist) to expose operators to clients.
4. **Counts**: inPartition two-pass; column metadata in `DataSetResponse.columns`
   (searchable/orderable flags already derivable from metadata + Config).

## SQL safety

- Column/target names are whitelisted against `query.Columns` — a `field:` that
  doesn't resolve never reaches SQL as an identifier.
- All values are bound parameters (`WhereExpr` `?` placeholders renumbered at
  Build) — no string interpolation of user input.

## Downstream adoption (redline2 `internal/dataset`)

Today redline2 does a placeholder free-text search (OR of ILIKE over a hardcoded
`FieldConfig.Searchable`). When this lands:
- `dataset.Run` calls `b.ApplySearch(req.Search, cfg)` and `ApplySearch(req.Partition, …)`.
- redline2's `FieldConfig` maps to `search.Config` (searchable/orderable/global).
- The placeholder OR-ILIKE is deleted.
No redline2 API/UI change — `DataSetRequest.search`/`partition` already exist.

## Open questions

1. ~~Parser: Option A vs B?~~ **DECIDED: Option A** — ANTLR Go target from the
   shared `../dataset` `.g4` (ANTLR 4.13.2, `github.com/antlr4-go/antlr/v4`
   runtime, generated parser vendored). See "Parser strategy" above.
2. ~~Where does the parser package live?~~ **DONE: `metaquery/search`** — free
   `Apply(b, query, cfg)` keeps the antlr dep out of core metaquery (a Builder
   method would force the dep onto every metaquery user).
3. ~~Unknown `field:` target → error vs literal?~~ **DONE: global value fallback**
   (matches dataset's `SearchConditionFactory` runtime, which drops the target
   and searches the value across global columns).
4. Time-kind default (no operator) — eq on a date? **Deferred to phase 3** (time
   columns are currently not searchable by default; overridable via `Search`).
5. Conformance: `conformance_test.go` ports the Kotlin `SearchParser` test
   corpus verbatim. **Parity: all realistic inputs identical** (targeting,
   commas/spaces/groups, apostrophes/quotes, `!`/`!!a` negation+escape,
   searchRendered auto-escape). One documented divergence: a search of *only*
   `:` — Kotlin escapes it to `\:`; antlr4-go single-token-deletes it during
   Sync (no virtual dispatch to intercept + unexported ATN internals), yielding
   an empty search. Pathological + harmless. Fixes made for parity: override
   `RecoverInline` (Go silently inserts missing tokens where Kotlin throws), and
   use a bool escape-signal instead of a `pos<0` sentinel (a NoViableAlt at
   offset 0 legitimately yields pos -1).
