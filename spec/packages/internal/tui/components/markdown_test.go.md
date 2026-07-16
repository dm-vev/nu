# `internal/tui/components/markdown_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Tests Markdown styling and width behavior.

## TODO

- [x] Tests cover inline styles and block wrapping.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Keep Markdown support usable for TUI messages without pulling dependencies.

## Tests

### `TestMarkdownMarkdownRendersInlineStyles`

Logic:
- Render strong, emphasis, and code spans with style callbacks.

Acceptance:
- Output contains the configured ANSI sequences.

### `TestMarkdownMarkdownRendersBlocksAndWraps`

Logic:
- Render heading, list, quote, and fenced code content in a narrow width.

Acceptance:
- Plain output includes expected content and visible widths stay bounded.

### `TestMarkdownMarkdownRendersPipeTables`

Logic:
- Render a pipe table with a header, separator, body rows, and inline strong text inside a cell.

Acceptance:
- Plain output contains aligned rows, omits the separator row, and preserves inline cell styling.
