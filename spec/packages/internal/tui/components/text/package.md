# `internal/tui/components/text`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Cached wrapping text component with optional background fill.

## Files

- `options.go`: padding/background options.
- `text.go`: component state.
- `cache.go`: render cache.
- `background.go`: background application.
- `render.go`: wrap, margin, and pad.
- `text_test.go`: wrap/pad checks.

## Acceptance Criteria

- Empty text renders no lines.
- Cache invalidates on text/background changes.
- Visible width stays within the supplied width.
