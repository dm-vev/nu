# `internal/tui/components/fill`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Flexible blank component used by the engine to push later components, currently
the editor and footer, to the bottom of the terminal.

## Files

- `fill.go`: component state and fixed render behavior.
- `render.go`: assigned-height blank row rendering.
- `fill_test.go`: assigned row checks.

## Acceptance Criteria

- `Render` returns no fixed rows.
- `FillLines` returns exactly the number of assigned blank padded rows.
- The component never writes to the terminal.
