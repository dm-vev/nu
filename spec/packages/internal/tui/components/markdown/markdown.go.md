# `internal/tui/components/markdown/markdown.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Holds Markdown source and options.

## TODO

- [x] File exists in the split component architecture.
- [x] Component methods are covered by render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Expose Markdown as a reusable `core.Component`.

## Functions

### `func New(source string, opts Options) *Markdown`

Logic:
- Store source and normalized options.

Acceptance:
- Callers can create a renderable Markdown component in one call.

### `func (m *Markdown) SetText(source string)`

Logic:
- Replace the raw Markdown source.

Acceptance:
- Future incremental renderers can reuse the component.

### `func (m *Markdown) Text() string`

Logic:
- Return raw Markdown source.

Acceptance:
- Tests and parent components can inspect source if needed.

### `func (m *Markdown) Invalidate()`

Logic:
- Satisfy the invalidatable component convention.

Acceptance:
- Container invalidation can call it safely.
