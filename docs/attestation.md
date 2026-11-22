# Attestation: dev features D1, D2, D3, D5, D7, D9

This document attests the owed dev features of `sqlc-go-codegen-metaquery`
against the project's declared surface (what is in the Makefile and go.mod at
the base commit under attestation), not against what the surface ought to be.

Declared surface at `main@41add33bf`:

- **Check** (declared): `go vet ./...` (Makefile `build` is `go build ./...`;
  the declared check per the issue is `go vet ./...`).
- **Test suite** (declared): `go test ./...` (Makefile `test` builds the wasm
  first, then runs the suite).
- **No fmt stage, no gen stage** declared. `make antlr` is a manual
  regeneration, not a declared stage.
- **No setup command** declared. There is no `make setup` and nothing in the
  Makefile makes a fresh checkout ready.

## Verification evidence

Measured on the branch being attested. Base is `main@41add33bf`; the branch is
0 commits behind base.

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

_The per-rung HOLD / vacuous-hold / gap statements for D1, D2, D3, D5, D7, D9
follow below. They are written against the verification evidence above and the
declared surface at the top of this file._
