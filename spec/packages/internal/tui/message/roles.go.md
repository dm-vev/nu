# `internal/tui/message/roles.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Role constants define TUI chat ownership.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Role constants are used instead of ad-hoc strings in TUI app state.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Define stable role strings for user and assistant messages.

## Code Style

Keep this file as constants only.

## Acceptance Criteria

- `RoleUser` is the only user role consumed by TUI chat rebuild.
- `RoleAssistant` is the only assistant role consumed by TUI chat rebuild.

## Constants

### `const RoleUser`

Logic:
- Mark messages authored by the user.

Acceptance:
- Used by `internal/tui` instead of local duplicate constants.

### `const RoleAssistant`

Logic:
- Mark messages authored by the assistant.

Acceptance:
- Used by `internal/tui` instead of local duplicate constants.
