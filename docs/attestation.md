# Attestation: dev features D1, D2, D3, D5, D7, D9

This document attests the owed dev features of `sqlc-go-codegen-metaquery`
against the project's declared surface (what is in the Makefile and go.mod at
the base commit under attestation), not against what the surface ought to be.

Declared surface (unchanged between `main@41add33bf` and the current base
`main@cf8685d30`):

- **Check** (declared): `go vet ./...` (Makefile `build` is `go build ./...`;
  the declared check per the issue is `go vet ./...`).
- **Test suite** (declared): `go test ./...` (Makefile `test` builds the wasm
  first, then runs the suite).
- **No fmt stage, no gen stage** declared. `make antlr` is a manual
  regeneration, not a declared stage.
- **No setup command** declared. There is no `make setup` and nothing in the
  Makefile makes a fresh checkout ready.

## Verification evidence

Measured on the branch being attested. Originally measured against
`main@41add33bf`; re-verified on the current base
`main@cf8685d30644bff8cd07b12a0add7b8645d42fbc`, 0 commits ahead of the cut
point and 0 of this branch's own. The re-verification at the current base:

- Declared check `go vet ./...`: exit 0, no output — PASS.
- Declared test suite as a whole `go test ./...`: exit 0, 140 pass / 0 fail /
  0 skip. The count moved from 135 at `main@41add33bf` to 140 at
  `main@cf8685d30` as tests landed on base; the per-package results are
  unchanged in shape (all packages `ok`).
- Name-filtered run `go test -run TestGenerate_RegenTestdataStock ./...`:
  exit 0, 1 pass / 0 fail / 0 skip; every other package reports
  `[no tests to run]`, confirming the declared suite filters by test name.

The detail below is from the original measurement at `main@41add33bf`; it
remains accurate for the declared surface, which has not changed between the
two base commits.

### Declared check: `go vet ./...`

- Command: `go vet ./...`
- Exit status: 0
- Recorded against: `main@41add33bfa3190a5853d6469b2e455588a2c2886`
- Result: PASS (no output)

### Declared test suite, as a whole: `go test ./...`

- Command: `go test ./...`
- Exit status: 0
- Duration: ~0.9 s
- Total tests: 135
- Pass: 135
- Fail: 0
- Skip: 0

Packages exercised (with test files):

| Package | Result |
|---|---|
| `github.com/iodesystems/sqlc-go-codegen-metaquery/internal` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/internal/opts` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/internal/endtoend` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/dataset` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqpgx` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/mqsqlite` | ok |
| `github.com/iodesystems/sqlc-go-codegen-metaquery/metaquery/search` | ok |

Packages with no test files (reported as `?` by `go test`, not counted in the
135): `internal/debug`, `internal/inflection`, `metaquery/codegen`, `plugin`.

Per-test lines (one per test, `--- PASS: <Name> (<time>)`): see the run's
`output.log` under the test-harness output directory. All 135 lines read
`--- PASS`. No `--- FAIL`, no `--- SKIP`.

### Declared test suite, name-filtered: `go test -run TestGenerate_RegenTestdataStock ./...`

- Command: `go test -run TestGenerate_RegenTestdataStock ./...`
- Exit status: 0
- Duration: ~0.6 s
- Tests run: 1 (only `TestGenerate_RegenTestdataStock`)
- Pass: 1
- Fail: 0
- Skip: 0

Every package reports `[no tests to run]` except
`github.com/iodesystems/sqlc-go-codegen-metaquery/internal/endtoend`, which
runs and passes `TestGenerate_RegenTestdataStock` in 0.00 s. This is the
observable evidence that the declared suite filters by test name and that
only matching tests run.

## Per-rung attestation

Statements are written against the declared surface at the top of this file
and the verification evidence above. They attest what the project **declares**,
not what it ought to declare. The attestation events (`feature.defined`,
`feature.attest`, `test.reported`) are not in the dev toolset; this document is
the evidence and prose a reviewer renders those events from.

### D1 — It builds

**HOLD.**

- The spec needs: a declared check, and that check exiting 0 on the committed
  tree.
- The project declares one: `go vet ./...` (the issue's declared check; the
  Makefile `build` target is `go build ./...`).
- The check ran on the committed tree and exited 0. The full declared test
  suite also runs clean (140 pass / 0 fail / 0 skip at the current base; 135 at
  the original base).

Nothing to fail, nothing to caveat. This is the cleanest rung in the attestation.

### D2 — The test library

**HOLD at the tool level, with a named gap at the dev-toolset level.**

- The spec needs: a suite that runs as a whole, that filters by test name, and
  that reports per-test results.
- All three are met by what the project declares (`go test ./...`) and by what
  the verification evidence shows:
  - as a whole: 140 pass / 0 fail / 0 skip (re-verified at the current base);
  - name-filtered: `go test -run TestGenerate_RegenTestdataStock ./...` runs
    exactly that one test and reports the rest of the packages as
    `[no tests to run]`;
  - per-test results: each run emits one `--- PASS: <Name> (<time>)` line per
    test, readable under the test-harness output directory.
- **Named gap (not attested):** "list the tests without running them" and
  "run the declared check as a toolset primitive" have no dev-toolset
  counterpart. Listing is available only via `go test -list`, which the dev
  toolset does not expose; the declared check is run as part of the project
  check, not through a dedicated primitive. The suite's listability is recorded
  here as a gap, not as attested.

### D3 — Every promise names what proves it

**HOLD at the aw4 level.**

- The spec needs: attribution from spec to tests, refusal to accept an
  attributed test that is not listable, and a view of specs with no proof yet.
- All three are aw4 capabilities, exercised in this work: the diff tool
  lists each spec id claimed and the tests attributed to it (or the fact that
  nothing proves it yet), the harness refuses unlistable attributed tests, and
  the session brief carries the unproven-spec view ("NOTHING PROVES THIS" per
  spec).
- The dev toolset does not carry the attribution tools; this is an aw4-level
  statement, and the reviewer renders the attestation events from this prose.

### D5 — The change is visible

**HOLD at the aw4 level.**

- The spec needs: the change readable as a diff against its base, and the
  specs a change claims to deliver named in it.
- Both are aw4 session capabilities: `diff` shows committed, uncommitted and
  untracked work against the branch's base, and accepts `covered_by` spec ids
  which are listed with their attributed tests.
- As with D3, the statement is at the aw4 level; the dev toolset does not carry
  the diff-reading tools.

### D7 — The tree is generated, formatted, linted and tested before hand-off

**HOLD vacuously.** This is the rung a reviewer is most likely to misread, so
the vacuousness is spelled out.

- The spec's scenarios are conditional: "given a project that declares a fmt
  stage …" and "given a project that declares a gen stage …".
- The project declares **no fmt stage and no gen stage**. There is no
  `make fmt` target and no `make gen` target; `make antlr` is a manual
  regeneration, not a declared stage.
- Both scenarios' antecedents are therefore false: there is nothing declared to
  be formatted, nothing declared to be generated, and so nothing that could
  fail to be formatted or generated before hand-off.
- The tested half of the rung is not vacuous: the declared test suite is run
  and clean (see D1 / D2).
- **This is not "the tree is formatted."** It is "nothing is declared to be
  formatted, so the fmt scenario has nothing to check." A reviewer reading this
  rung as evidence of formatting hygiene would be reading it wrong.

### D9 — A fresh checkout is made ready

**NOT MET AS DECLARED.** This is the negative rung, and it is the most useful
one in the attestation.

- The spec needs: a project that declares a setup command, and that a fresh
  checkout running it once is made ready.
- The project declares **no setup command**. There is no `make setup` target
  and nothing in the Makefile makes a fresh checkout ready.
- The scenario's antecedent ("given a project that declares a setup command")
  is false for this project, so spec 1 of D9 is not met as-declared. This is
  not a vacuous-hold in the D7 sense: the dev tree owes D9 to this project
  (it has the `Filesystem` property class and no exemption applies), so the
  rung is owed and not met, not unowed.
- **Candidate fixes (PM follow-up, not resolved here):**
  1. Declare a setup command — e.g. add `make setup` = `go mod download` (or
     whatever the project's fresh-checkout step is) to the Makefile, and re-run
     the attestation against the new declared surface.
  2. Record D9 as not-applicable for this project, with the reason (a Go
     project whose only external dependency is the Go toolchain and the module
     cache needs no separate setup step beyond `go mod download`, which the
     Go toolchain performs on demand).
- Until one of these is done, the attestation records D9 as a gap.

## Summary

| Rung | Statement |
|---|---|
| D1 | HOLD — declared check `go vet ./...` exits 0 on the committed tree |
| D2 | HOLD at tool level; named gap: no dev-toolset primitive for list-without-running or running the declared check directly |
| D3 | HOLD at aw4 level — attribution, refusal, and unproven-spec view are aw4 capabilities |
| D5 | HOLD at aw4 level — diff-reading and spec-coverage naming are aw4 capabilities |
| D7 | HOLD vacuously — no fmt/gen stage declared; the tested half is met |
| D9 | NOT MET AS DECLARED — no setup command declared; candidate fixes recorded as PM follow-up |
