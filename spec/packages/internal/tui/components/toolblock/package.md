# `internal/tui/components/toolblock`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Render command, patch, and generic tool executions as separate Pi-like blocks.
The package chooses pending/success/error backgrounds and extracts useful text
from built-in tool JSON results.

## Files

- `options.go`: block background and text/diff style callbacks.
- `block.go`: tool execution state.
- `render.go`: box composition.
- `display.go`: command/result extraction and formatting.
- `diff.go`: patch/diff line coloring.
- `toolblock_test.go`: command, failure, and patch rendering tests.

## Acceptance Criteria

- Pending, success, and error states can render with different backgrounds.
- Bash results show `$ command` and output.
- Edit results show patch lines with added/removed/context colors.
- `go test ./internal/tui/components/toolblock` passes.
