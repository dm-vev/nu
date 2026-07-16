# `internal/tui/markdown_options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines Markdown padding and style callbacks.

## TODO

- [x] File exists in the split component architecture.
- [x] Style defaults are identity-safe.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Configure Markdown render styling without importing application theme code.

## Acceptance Criteria

- Negative padding normalizes to zero.
- Missing style functions default to visible unstyled output.

## Types

### `type MarkdownOptions struct`

Logic:
- Carry padding plus text, heading, strong, emphasis, code, quote, and bullet styles.

Acceptance:
- Component users can provide app-specific colors without changing this package.

## Functions

### `func markdownNormalizeOptions(opts MarkdownOptions) MarkdownOptions`

Logic:
- Clamp padding and fill nil style callbacks.

Acceptance:
- Rendering never panics because a style callback is nil.

### `func markdownIdentity(value string) string`

Logic:
- Return text unchanged.

Acceptance:
- Used as the default style callback.
