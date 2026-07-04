# `internal/tui/components/fill/render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Render blank rows for a height assigned by `engine.TUI`.

## Functions

### `func (f *Fill) FillLines(width int, rows int) []string`

Logic:
- Return `rows` blank strings padded to `width`; return no lines for zero or negative rows.

Acceptance:
- Covered by `go test ./internal/tui/components/fill`.
