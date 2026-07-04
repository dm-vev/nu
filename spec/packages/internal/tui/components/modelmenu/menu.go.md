# `internal/tui/components/modelmenu/menu.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Selector state and public component API.

## Purpose

Own visible state, current model identity, and selected model lookup.

## Functions

### `New(models []model.Model, opts Options) *Menu`

Logic:
- Copy candidates, normalize options, and start hidden.

Acceptance:
- Construction performs no provider/auth/terminal work.

### `(*Menu) SetModels(models []model.Model)`

Logic:
- Replace candidates with a copy and refresh filtering when visible.

Acceptance:
- Runtime can refresh candidates before opening `/model`.

### `(*Menu) Open(query string, currentProvider string, currentID string)`

Logic:
- Store query and current identity, mark visible, and refresh filtered rows.

Acceptance:
- Current model is marked and sorted first.

### `(*Menu) Close()`

Logic:
- Hide the selector and clear transient query/filter/selection state.

Acceptance:
- Cancel returns the editor to normal input behavior.

### `(*Menu) Visible() bool`

Logic:
- Report whether selector input/render paths are active.

Acceptance:
- TUI input can route events without inspecting selector internals.

### `(*Menu) Query() string`

Logic:
- Return current search text.

Acceptance:
- Tests can verify search input mutations.

### `(*Menu) Selected() (model.Model, bool)`

Logic:
- Return highlighted model when the filtered list has one.

Acceptance:
- Enter can select without duplicating index logic.

### `(*Menu) Invalidate()`

Logic:
- Satisfy component invalidation convention.

Acceptance:
- Containers can invalidate the selector safely.
