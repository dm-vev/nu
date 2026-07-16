# `internal/tui/editor_state.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Store editor buffer lines and cursor position.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Store editor buffer lines and cursor position.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type State struct {`

Logic:
- State stores editor buffer and cursor.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func editorInitialState() State`

Logic:
- Create the initial one-line empty buffer with cursor at row 0, column 0.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
