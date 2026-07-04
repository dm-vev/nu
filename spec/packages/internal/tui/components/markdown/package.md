# `internal/tui/components/markdown`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Small stdlib-only Markdown renderer for terminal message content. It supports
the subset required by Pi-like chat output: headings, lists, block quotes,
fenced code, inline bold, inline italic, and inline code.

## Files

- `options.go`: style callbacks and padding.
- `markdown.go`: component state and mutators.
- `render.go`: padding, wrapping, and width bounding.
- `block.go`: block-level Markdown parsing.
- `inline.go`: inline marker parsing.
- `markdown_test.go`: Markdown style and wrapping tests.

## Acceptance Criteria

- Empty Markdown renders no lines.
- Styled visible width stays within render width.
- Inline code/bold/italic produce ANSI-styled segments when configured.
- `go test ./internal/tui/components/markdown` passes.
