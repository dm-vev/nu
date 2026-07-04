# `internal/slash/commands_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Tests

### `TestBuiltinsCopiesPiCommandSet`

Logic:
- Assert the built-in command names match Pi order.

Acceptance:
- Fails on missing, extra, or reordered commands.

### `TestParseSlashCommand`

Logic:
- Parse command name and arguments from one slash input.

Acceptance:
- Command dispatch receives stable name/args fields.

### `TestFilterSlashCommands`

Logic:
- Filter `/mo` and expect model-related commands.

Acceptance:
- Menu filtering remains useful for partial input.
