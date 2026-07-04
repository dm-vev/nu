# `internal/tui/engine/cursor.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define cursor marker and extracted cursor position data.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define cursor marker and extracted cursor position data.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type cursorPosition struct {`

Logic:
- Store whether a cursor marker was found and the row/column where the hardware cursor should move.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func extractCursor(lines []string, rows int) cursorPosition`

Logic:
- Scan rendered lines for `core.CursorMarker`, remove it, and compute visible row/column within the terminal viewport.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *TUI) positionCursor(cursor cursorPosition) error`

Logic:
- Move the terminal cursor from the engine's current row to the extracted cursor row and update cached cursor coordinates.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
