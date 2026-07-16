# `internal/tui/markdown_render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Renders parsed Markdown blocks within terminal width.

## TODO

- [x] File exists in the split component architecture.
- [x] Wrapping is covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Convert parsed Markdown lines into padded, wrapped terminal lines.

## Acceptance Criteria

- Empty source returns nil.
- ANSI-stripped line widths never exceed the supplied width.
- Blank lines inside content are preserved after trimming leading/trailing blanks.

## Functions

### `func (m *Markdown) Render(width int) []string`

Logic:
- Parse block lines, trim blank edges, wrap by visible width, and apply padding.

Acceptance:
- `go test ./internal/tui` covers wrapping and styles.

### `func markdownTrimBlankEdges(lines []string) []string`

Logic:
- Remove leading and trailing blank lines while preserving internal separation.

Acceptance:
- Markdown blocks do not introduce stray top/bottom whitespace.
