# `internal/slash/command.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Parse and dispatch slash commands.

## Code Style

Parsing is pure. Handlers live in `builtin.go` or extension registry.

## Functions

### `Parse(input string) (Invocation, bool)`

Logic:

- Inspect only the beginning of the editor buffer; leading non-command text means `ok=false`.
- Treat `/name` followed by whitespace or end-of-input as a command name.
- Keep the raw argument tail byte-for-byte so prompt templates and extensions can parse their own syntax.
- Do not expand files, environment, aliases, or prompt templates in the parser.

Acceptance:

- detects slash command at start of editor input;
- preserves raw args after command name.

### `Dispatch(ctx context.Context, inv Invocation, reg Registry) error`

Logic:

- Check context before resolving the command.
- Resolve command name by precedence: built-in, extension command, skill shortcut, then prompt template.
- Pass raw args and caller services to the selected handler without reparsing globally.
- Return a typed unknown-command error that TUI can render without provider involvement.

Acceptance:

- resolves built-in, extension, skill, and prompt-template commands.

Tests:

- `TestNUF110SlashCommandDispatch`
- `TestNUF110UnknownCommandReportsError`
