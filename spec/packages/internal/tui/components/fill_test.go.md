# `internal/tui/components/fill_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Verify fixed render behavior and assigned-height blank rows.

## Functions

### `func TestFillFillRendersAssignedRows(t *testing.T)`

Logic:
- Assert `Render` has no fixed rows and `FillLines` returns exactly assigned blank rows.

Acceptance:
- Fails if fill stops anchoring later components by consuming fixed height.
