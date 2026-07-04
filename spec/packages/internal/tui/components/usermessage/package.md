# `internal/tui/components/usermessage`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

User turn block with prompt-zone OSC 133 markers, padded background, and
Markdown rendering.

## Files

- `options.go`: text, Markdown, background styles, and padding.
- `message.go`: raw user message state.
- `render.go`: box/Markdown composition and OSC markers.
- `message_test.go`: text and marker checks.

## Acceptance Criteria

- Rendered user turns include OSC 133 start/end/final markers.
- Markdown text is preserved after styling.
- User message background fills every rendered line.
