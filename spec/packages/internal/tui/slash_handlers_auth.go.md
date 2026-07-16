# `internal/tui/tui_slash_handlers_auth.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Trust/login/logout command behavior was split from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Handle local project trust and provider credential slash commands.

## Code Style

Use injected home/cwd paths, never call providers, and never echo stored secrets.

## Owned Logic

- `/trust` stores a cwd boolean in the global trust file.
- `/login` stores direct, environment-reference, or command-reference credentials.
- `/logout` removes one provider while preserving others.
- Path helpers derive auth/trust files from injected home with cwd fallback.

## Acceptance

- Tests never access real `~/.nu`.
- Usage errors stay local and file failures are returned.
- Login/logout preserve unrelated providers.

## Tests

- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
- `TestTUISlashSessionDoesNotCallAgent`
