# `internal/tui/model_menu_modelmenu_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Model selector unit tests.

## Purpose

Cover current-first rendering, display-name filtering, and Enter selection action.

## Tests

### `TestModelMenuModelMenuRendersCurrentModelFirst`

Acceptance:
- Current model is first and marked.

### `TestModelMenuModelMenuFiltersByDisplayNameAndSelects`

Acceptance:
- Display-name search finds the expected model and Enter returns `ModelMenuActionSelect`.
