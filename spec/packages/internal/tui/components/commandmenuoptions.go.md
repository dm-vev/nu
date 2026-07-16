# `internal/tui/components/commandmenuoptions.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Configure command menu row count and style callbacks.

## Functions

### `commandMenuNormalizeOptions(opts CommandMenuOptions) CommandMenuOptions`

Logic:
- Default max rows and nil style callbacks.

Acceptance:
- Menu renders with zero-value options.

### `commandMenuIdentity(value string) string`

Logic:
- Return text unchanged for default styling.

Acceptance:
- Tests can render without ANSI callbacks.
