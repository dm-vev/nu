# `internal/tui/core/line.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Apply per-line terminal resets.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Apply per-line terminal resets.

## Code Style

Keep interfaces tiny and structural. Containers own ordering only, not layout policy. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.
## Functions

### `func ResetLines(lines []string) []string`

Logic:
- ResetLines appends Pi-compatible resets to non-image lines.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
