# `internal/tui/components/spacer`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Blank-line spacer component.

## Files

- `spacer.go`: line count state.
- `render.go`: blank padded lines.
- `spacer_test.go`: line count check.

## Acceptance Criteria

- Negative counts clamp to zero.
- Rendered blank lines match requested width.
