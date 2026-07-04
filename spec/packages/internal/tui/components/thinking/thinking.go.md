# `internal/tui/components/thinking/thinking.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Holds model reasoning source.

## TODO

- [x] File exists in the split component architecture.
- [x] State methods are covered by render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Expose reasoning content as a reusable TUI component.

## Functions

### `func New(value string, opts Options) *Thinking`

Logic:
- Store reasoning text and normalized options.

Acceptance:
- Assistant message can construct thinking blocks directly.

### `func (t *Thinking) SetText(value string)`

Logic:
- Replace reasoning source.

Acceptance:
- Future incremental renderers can reuse the component.

### `func (t *Thinking) Text() string`

Logic:
- Return raw reasoning source.

Acceptance:
- Tests can inspect component state when needed.

### `func (t *Thinking) Invalidate()`

Logic:
- Satisfy the invalidatable component convention.

Acceptance:
- Container invalidation can call it safely.
