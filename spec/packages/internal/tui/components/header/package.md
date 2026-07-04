# `internal/tui/components/header`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Startup header with Nu logo, compact key hints, expanded help, and onboarding
text.

## Files

- `options.go`: style callbacks and padding.
- `header.go`: state and expansion methods.
- `content.go`: compact/expanded text builders.
- `render.go`: wrapping and padding.
- `header_test.go`: compact/expanded checks.

## Acceptance Criteria

- `ctrl+o` toggles compact/expanded content through the top-level app.
- No header line exceeds the supplied width.
