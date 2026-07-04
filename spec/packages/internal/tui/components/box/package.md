# `internal/tui/components/box`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Padding and background container around child components.

## Files

- `options.go`: padding/background options.
- `box.go`: child list operations.
- `background.go`: background fill.
- `render.go`: padded child rendering.
- `box_test.go`: child padding check.

## Acceptance Criteria

- Empty boxes render no lines.
- Children render at content width.
- Output lines are padded to the requested width.
