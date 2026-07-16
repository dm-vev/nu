# `internal/tui/components/thinking_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Tests reasoning style output.

## TODO

- [x] Test validates italic and dim ANSI markers.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Prevent thinking blocks from regressing into normal assistant text.

## Tests

### `TestThinkingThinkingRendersMarkdownWithThinkingStyle`

Logic:
- Render Markdown with thinking style callbacks.

Acceptance:
- Output contains dim/italic ANSI and visible reasoning text.
