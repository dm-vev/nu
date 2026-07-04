# `internal/tui/components/modelmenu/input.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Raw input handling for the model selector.

## Purpose

Translate terminal events into selector actions.

## Functions

### `(*Menu) HandleInput(data string) Action`

Logic:
- Up/down arrows wrap selection.
- Enter returns select, Escape/Ctrl+C returns cancel.
- Backspace edits query.
- Printable text appends to query and refreshes filtering.

Acceptance:
- Selector input never mutates the main editor while visible.

### `(*Menu) move(delta int)`

Logic:
- Move selected index with wraparound and no-op on empty lists.

Acceptance:
- Up/down can cycle all visible candidates.

### `(*Menu) backspace()`

Logic:
- Remove one rune from query and refresh filtering.

Acceptance:
- UTF-8 query text is edited by rune, not byte.

### `isPrintable(data string) bool`

Logic:
- Accept non-control text input and reject terminal control sequences.

Acceptance:
- Arrow/Escape/control bytes are not inserted into selector search.
