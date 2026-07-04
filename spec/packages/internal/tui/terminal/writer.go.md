# `internal/tui/terminal/writer.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Write terminal bytes and helper control sequences.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Write terminal bytes and helper control sequences.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.
## Functions

### `func (t *Terminal) Write(data string) error`

Logic:
- Write writes raw terminal bytes.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) HideCursor() error`

Logic:
- HideCursor hides the hardware cursor.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) ShowCursor() error`

Logic:
- ShowCursor shows the hardware cursor.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) MoveBy(lines int) error`

Logic:
- MoveBy moves cursor vertically.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) SetTitle(title string) error`

Logic:
- SetTitle sets terminal title.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
