# `internal/tui/message/message.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Constructors keep top-level TUI state concise.

## TODO

- [x] File exists in the split `internal/tui/message` architecture.
- [x] Constructors are covered by package tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Create user/assistant messages and clone message part slices safely.

## Code Style

Constructors must not allocate more structure than needed for the current role.

## Acceptance Criteria

- User messages start with one text part.
- Assistant messages may start empty for streamed content.
- Clone returns independent part storage.

## Functions

### `func NewUser(text string) Message`

Logic:
- Return a `RoleUser` message containing a single `PartText`.

Acceptance:
- Submit path stores user prompts through this constructor.

### `func NewAssistant() Message`

Logic:
- Return an empty `RoleAssistant` message for streamed parts.

Acceptance:
- Tool/thinking/text mutations can build the message incrementally.

### `func NewAssistantText(text string) Message`

Logic:
- Create an assistant message and append one text part.

Acceptance:
- Error/final assistant messages can be created in one call.

### `func (m Message) Clone() Message`

Logic:
- Copy the message and its part slice.

Acceptance:
- Callers can mutate the returned message without aliasing the original parts.
