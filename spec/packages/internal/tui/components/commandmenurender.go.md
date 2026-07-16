# `internal/tui/components/commandmenurender.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Render command suggestions as padded terminal rows.

## Functions

### `(*CommandMenu) Render(width int) []string`

Logic:
- Return no rows when hidden.
- Render matching commands with the highlighted row selected.
- Render `No matching commands` for slash prefixes with no match.

Acceptance:
- Output visible width stays within terminal width.

### `commandMenuMenuPrefix(text string) (string, bool)`

Logic:
- Accept only a single slash-command prefix without spaces.

Acceptance:
- Command menu does not appear while typing normal prompts or command arguments.
