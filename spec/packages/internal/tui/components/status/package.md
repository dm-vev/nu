# `internal/tui/components/status`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Single transient status line for working/tool/aborting states.

## Files

- `status.go`: text state and style callback.
- `render.go`: idle/non-empty rendering.
- `status_test.go`: idle and busy checks.

## Acceptance Criteria

- Empty status renders no lines.
- Non-empty status renders one padded line.
