# `internal/tui/components/markdowninline.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Parses a small inline Markdown subset.

## TODO

- [x] File exists in the split component architecture.
- [x] Bold, italic, and inline code are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Style inline Markdown markers without adding a parsing dependency.

## Functions

### `func markdownRenderInline(source string, opts MarkdownOptions) string`

Logic:
- Scan source left to right and style `**strong**`, `*emphasis*`, `_emphasis_`,
  and `` `code` `` spans.
- Leave unmatched markers visible as text.

Acceptance:
- Plain text, strong, emphasis, and code spans can coexist in one line.

### `func markdownNextMarker(source string, start int) int`

Logic:
- Find the next candidate inline marker after a plain-text segment.

Acceptance:
- Plain text is emitted in the largest safe segment.
