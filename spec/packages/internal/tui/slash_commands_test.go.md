# `internal/tui/slash_commands_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Tests

### `TestSlashBuiltinsCopiesPiCommandSet`

Logic:
- Assert the built-in command names match Pi order.

Acceptance:
- Fails on missing, extra, or reordered commands.

### `TestSlashParseSlashCommand`

Logic:
- Parse command name and arguments from one slash input.

Acceptance:
- Command dispatch receives stable name/args fields.

### `TestSlashFilterSlashCommands`

Logic:
- Filter `/mo` and expect model-related commands.

Acceptance:
- Menu filtering remains useful for partial input.
