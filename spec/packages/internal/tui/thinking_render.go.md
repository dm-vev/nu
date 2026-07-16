# `internal/tui/thinking_render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Delegates to Markdown with thinking styles.

## TODO

- [x] File exists in the split component architecture.
- [x] Markdown delegation is covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Render reasoning as gray/italic Markdown.

## Functions

### `func (t *Thinking) Render(width int) []string`

Logic:
- Create a Markdown component with thinking styles for text, headings, quote,
  bullets, strong, emphasis, and code.

Acceptance:
- Output preserves Markdown behavior while visually separating reasoning.
