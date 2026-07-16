# `internal/tui/tool_block_options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines tool block theme hooks.

## TODO

- [x] File exists in the split component architecture.
- [x] Nil text callbacks normalize safely.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Keep tool block styling injectable from the app palette.

## Functions

### `func toolBlockNormalizeOptions(opts ToolBlockOptions) ToolBlockOptions`

Logic:
- Clamp padding and default nil style callbacks.

Acceptance:
- Tool blocks can render with partial options during tests.

### `func toolBlockIdentity(value string) string`

Logic:
- Return text unchanged.

Acceptance:
- Used for unset style callbacks.
