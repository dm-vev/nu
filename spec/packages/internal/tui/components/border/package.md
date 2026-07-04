# `internal/tui/components/border`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Horizontal border component used by Pi-like surfaces.

## Files

- `border.go`: border state.
- `render.go`: width-sized line rendering.
- `border_test.go`: render check.

## Acceptance Criteria

- Width is clamped to at least one cell.
- Optional style wraps the full border line.
