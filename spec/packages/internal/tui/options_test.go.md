# `internal/tui/options_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Verify terminal charset detection for TUI fallback rendering.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Verify terminal charset detection for TUI fallback rendering.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.

## Functions

### `func TestLimitedCharsetDetectsOptionEnvAndTerm(t *testing.T)`

Logic:
- Verify explicit `AppOptions.ASCII`, truthy `NU_TUI_ASCII`, and limited `TERM` values enable ASCII rendering.
- Verify modern terminals keep full Unicode rendering by default.

Acceptance:
- The test fails if PicoCalc/Linux-console compatible fallback detection regresses.
