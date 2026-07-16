# `internal/tui/components/commandmenu_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Tests

### `TestCommandMenuMenuRendersFilteredCommands`

Logic:
- Render `/mo` suggestions and verify completion returns `/model `.

Acceptance:
- Menu displays matching Pi commands and drives Tab completion.

### `TestCommandMenuMenuHidesOutsideSlashPrefix`

Logic:
- Feed ordinary text containing `/mo`.

Acceptance:
- Menu stays hidden outside a leading slash command.

### `TestCommandMenuMenuMovesSelectionAndCompletesSelectedCommand`

Logic:
- Open the root slash menu, move selection down, and verify completion/render target the moved row.

Acceptance:
- The test fails if selection stays pinned to the first command.
