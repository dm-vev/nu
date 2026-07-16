# `internal/tui/fill_fill.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Declare the flexible blank component used for bottom anchoring.

## Functions

### `func NewFill() *Fill`

Logic:
- Return an empty flexible component.

Acceptance:
- Covered by `go test ./internal/tui`.

### `func (f *Fill) Render(width int) []string`

Logic:
- Return no fixed lines because height is assigned later by the engine.

Acceptance:
- Fixed component layout counts this component as zero rows before filler allocation.
