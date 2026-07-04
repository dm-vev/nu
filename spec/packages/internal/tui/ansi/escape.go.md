# `internal/tui/ansi/escape.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Parse escape sequence boundaries and define the per-line reset suffix.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Parse escape sequence boundaries and define the per-line reset suffix.

## Code Style

Keep helpers pure, allocation-light, and independent from component state. Never count escape bytes as visible cells. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Types And Constants

### `const ResetSuffix = "\x1b[0m\x1b]8;;\x07"`

Logic:
- ResetSuffix matches Pi's per-line reset.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func End(text string, start int) int`

Logic:
- End returns the byte index after the escape sequence at start.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
