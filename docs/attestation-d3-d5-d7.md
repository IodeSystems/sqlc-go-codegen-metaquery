# Attestation: D3, D5, D7

Base: main@41add33bf (measured 2026-11-21)
Project: sqlc-go-codegen-metaquery

This document attests three owed dev rungs against the project's declared surface.
Each rung names its specs, the evidence that proves them, and the basis (aw4-level
or vacuous) so a reviewer's `feature.attest` is grounded in measured fact.

---

## D3 — Every promise names what proves it

**Feature:** f_1m7yerp3d4hkthwz61aqh1

**Specs:**

| Spec ID | Scenario |
|---------|----------|
| sp_1m7yerp4a4xqterbkwjd3s | reviewer attributes a test to a spec |
| sp_1m7yerp4a0bcf5ytr78qtn | unlistable attributed test is refused |
| sp_1m7yerp4a42k2fk3wmgjq1 | tree shows unproven specs |

**Evidence:**

- Declared `test_suite` property: `argv=[go test ./...]`, `list=[go test ./... -list .*]`,
  `timeout_seconds=600`. The `list` command makes the suite listable, which is the
  prerequisite for attribution checks.
- Measured at base main@41add33bf: `go test ./...` = **135 pass, 0 fail, 0 skipped**
  (this session, exit 0, 1.4s).
- F1–F6 project specs carry attributed tests (8 closed subtasks under
  p_1m7kg5q0d7enzcw5ec6s60). The specs are attributable and listable.

**Basis: HOLDS at the aw4 level.**
The three D3 specs are properties of aw4's `feature.attest` / attribution machinery,
not of this project's tree. The project's declared `test_suite` property with its
`list` command provides the listable surface that the machinery operates on. The
135 measured tests confirm the suite is both runnable and listable.

---

## D5 — The change is visible

**Feature:** f_1m7yerpws05x2k6jw3nygq

**Specs:**

| Spec ID | Scenario |
|---------|----------|
| sp_1m7yerpy0dx0pmrm9v1316 | session reads the diff of work it is judging |
| sp_1m7yerpy0d7d1e76pjnebg | diff names covering specs |

**Evidence:**

- The aw4 `diff` tool is in the dev toolset. It shows changed files against the
  branch base (committed and uncommitted, including files added and never committed),
  and accepts a `covered_by` parameter listing spec ids whose attributed tests cover
  the changed lines.
- This tool is exercised in every dev session: the `diff` tool call returns the
  shape and content of the change with spec coverage.

**Basis: HOLDS at the aw4 level.**
Both D5 specs are properties of the aw4 `diff` tool itself. The tool exists in the
dev toolset, shows the full diff (not only touched files), and names which specs'
attributed tests cover the changed lines. No project-specific surface is required
beyond the attribution machinery that D3 already attests.

---

## D7 — The tree is generated, formatted, linted and tested before hand-off

**Feature:** f_1m7yerqr00xg8a5y0wn2h7

**Specs:**

| Spec ID | Scenario |
|---------|----------|
| sp_1m7yerqrx1p5w1a4km1jbp | fmt stage changes nothing on base |
| sp_1m7yerqrxbxspt6wg5pyk8 | gen stage changes nothing on base |

**Evidence:**

- The project's declared property surface contains: `filesystem`, `scm`,
  `test_suite`, `dev_server`, `check`. There is **no `fmt` class** and **no `gen`
  class** declared.
- Both D7 specs are conditional: "Given a project that declares a fmt stage" and
  "Given a project that declares a gen stage." With neither stage declared, the
  precondition is never met.

**Basis: HOLDS vacuously.**
No fmt stage is declared, so the fmt spec's precondition is never satisfied.
No gen stage is declared, so the gen spec's precondition is never satisfied.
Both specs are vacuously true against the declared surface.

Note: the Makefile has an `antlr` target, but the attestation is against the
declared property surface (project_properties), not the Makefile. No gen stage
is declared in that surface.

---

## Summary

| Rung | Feature | Verdict | Basis |
|------|---------|---------|-------|
| D3 | f_1m7yerp3d4hkthwz61aqh1 | HOLDS | aw4-level: attribution machinery + listable 135-test suite |
| D5 | f_1m7yerpws05x2k6jw3nygq | HOLDS | aw4-level: diff tool in dev toolset |
| D7 | f_1m7yerqr00xg8a5y0wn2h7 | HOLDS vacuously | no fmt/gen stage declared |
