# `internal/tui/slash_commands.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-compatible built-in slash command list and tiny parsing helpers.

## TODO

- [x] File exists.
- [x] Built-in command list is covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Expose the command list shared by TUI command menu and local slash dispatch.

## Functions

### `SlashBuiltins() []SlashCommand`

Logic:
- Return a copy of the Pi built-in command list.

Acceptance:
- `TestSlashBuiltinsCopiesPiCommandSet` fails if the list drifts.

### `SlashLookup(name string) (SlashCommand, bool)`

Logic:
- Trim a leading slash and find a built-in by exact name.

Acceptance:
- Unknown slash commands can be rejected before they reach the model.

### `SlashParse(input string) (string, string, bool)`

Logic:
- Parse `/name args` into name and args.

Acceptance:
- Non-slash text is not treated as a command.

### `SlashFilter(prefix string, limit int) []SlashCommand`

Logic:
- Filter built-ins by prefix or substring and cap the result when `limit > 0`.

Acceptance:
- Command menu can render filtered suggestions for partial `/` input.
