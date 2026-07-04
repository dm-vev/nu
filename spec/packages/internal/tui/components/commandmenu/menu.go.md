# `internal/tui/components/commandmenu/menu.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Small slash command suggestion component.

## Purpose

Show Pi-style slash command suggestions above the editor.

## Functions

### `New(commands []slash.Command, opts Options) *Menu`

Logic:
- Copy commands and normalize render options.

Acceptance:
- Construction performs no filesystem, terminal, or provider work.

### `(*Menu) SetText(text string)`

Logic:
- Show suggestions only when the editor contains a slash command prefix before the first space.
- Reset selection when the prefix changes and clamp it when the filtered list shrinks.

Acceptance:
- `TestMenuHidesOutsideSlashPrefix` fails if ordinary prompts show the menu.

### `(*Menu) Completion() (string, bool)`

Logic:
- Return the highlighted visible command as `/name `.

Acceptance:
- Tab completion can use the menu without duplicating filter logic.

### `(*Menu) Selected() (slash.Command, bool)`

Logic:
- Return the highlighted command when the menu has a valid selection.

Acceptance:
- Enter can execute the same command row that render highlights.

### `(*Menu) Move(delta int) bool`

Logic:
- Move the highlighted row with wraparound.

Acceptance:
- Arrow keys can move the command selector instead of changing editor text.

### `(*Menu) Visible() bool`

Logic:
- Report whether the command selector has visible matches.

Acceptance:
- Raw input routing can decide whether arrow/Enter belong to the menu.

### `(*Menu) Invalidate()`

Logic:
- Satisfy component invalidation convention.

Acceptance:
- Containers can invalidate the menu safely.
