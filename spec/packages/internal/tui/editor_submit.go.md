# `internal/tui/editor_submit.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Submit and clear editor text.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Submit and clear editor text.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func (e *Editor) submit()`

Logic:
- Read editor text, reset the buffer, call the submit callback only for non-empty trimmed text, and notify change listeners after clearing.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func editorJoinLines(lines []string) string`

Logic:
- Join editor buffer lines with `\n` for submit and `Text()`.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
