# `internal/tui/tool_block_block.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Stores one tool execution block.

## TODO

- [x] File exists in the split component architecture.
- [x] Block construction is covered by render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Represent the tool execution data required by the renderer.

## Functions

### `func NewToolBlock(toolName string, toolID string, arguments string, result string, state ToolState, opts ToolBlockOptions) *ToolBlock`

Logic:
- Store raw tool metadata, raw result, state, and normalized options.

Acceptance:
- Assistant message can create a block from one `Part`.

### `func (b *ToolBlock) Invalidate()`

Logic:
- Satisfy the invalidatable component convention.

Acceptance:
- Container invalidation can call it safely.
