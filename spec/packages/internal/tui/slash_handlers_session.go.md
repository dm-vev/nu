# `internal/tui/tui_slash_handlers_session.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Session/settings/import/tree/compact/reload handlers were split from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Handle slash commands that inspect or mutate current in-memory TUI session state.

## Code Style

Keep local-only behavior explicit, lock message state, and use shared import/export helpers for files.

## Owned Logic

- Render settings and linear tree tables.
- Import/append or resume/replace exported JSONL messages and rebuild rendered chat.
- Show/update names, read a local changelog, fork by index/latest user, and clone message values.
- Compact long local history to a marker plus six-message tail and reload git footer state.
- Helpers select fork indexes, escape table cells, and truncate rune-aware previews.

## Acceptance

- Commands do not call the agent and file paths stay under injected cwd resolution.
- Fork/clone/import update rendered state without aliasing source message values.
- Small histories are not compacted; long previews remain table-safe.

## Tests

- `TestTUISlashSessionDoesNotCallAgent`
- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
