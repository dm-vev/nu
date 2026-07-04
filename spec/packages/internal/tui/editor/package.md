# `internal/tui/editor`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Focused multiline editor component. It owns buffer state, rune-safe mutation,
submit/clear behavior, cursor marker insertion, and bordered rendering.

## Files

- `editor.go`: editor type and public methods.
- `state.go`: buffer lines and cursor position.
- `input.go`: decoded key/paste handling.
- `submit.go`: submit and joined text helpers.
- `render.go`: width-bounded editor rendering.
- `editor_test.go`: wrapping, submit, Unicode, paste, and delete tests.

## Acceptance Criteria

- Cursor positions are rune indexes, not byte indexes.
- Input mutation and rendering stay separate.
- Pasted newlines create editor lines.
- Rendered lines fit the requested width.
