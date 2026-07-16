# `internal/tui/components/toolblockdiff.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Colors patch lines without external diff dependencies.

## TODO

- [x] File exists in the split component architecture.
- [x] Added/removed/context lines are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Render built-in edit patches with green additions, red removals, and muted
context/header lines.

## Functions

### `func toolBlockRenderDiff(patch string, opts ToolBlockOptions) string`

Logic:
- Normalize line endings and tabs.
- Style `+` lines as additions, `-` lines as removals, and headers/context as muted context.

Acceptance:
- `+++` and `---` file headers are not treated as ordinary additions/removals.
