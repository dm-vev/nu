# `internal/tui/tool_block_render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Renders a tool execution as a padded box.

## TODO

- [x] File exists in the split component architecture.
- [x] Success/error backgrounds are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Compose formatted tool content with a background box.

## Functions

### `func (b *ToolBlock) Render(width int) []string`

Logic:
- Format content, choose pending/success/error background, and render it inside
  a padded box.

Acceptance:
- Failed command results get the error background even when the tool event state is success.
