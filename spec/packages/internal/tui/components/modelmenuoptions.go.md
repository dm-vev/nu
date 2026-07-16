# `internal/tui/components/modelmenuoptions.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Model selector render options.

## Purpose

Define styling and row-count options for the model selector.

## Functions

### `modelMenuNormalizeOptions(opts ModelMenuOptions) ModelMenuOptions`

Logic:
- Fill zero `MaxVisible` and nil style callbacks with deterministic defaults.

Acceptance:
- Tests can construct a menu with `ModelMenuOptions{}`.

### `modelMenuIdentity(value string) string`

Logic:
- Return the input string unchanged for default styling.

Acceptance:
- Nil style callbacks never panic during render.
