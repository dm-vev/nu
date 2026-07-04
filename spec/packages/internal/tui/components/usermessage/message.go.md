# `internal/tui/components/usermessage/message.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: User message component with prompt-zone markers and padded styling.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

User message component with prompt-zone markers and padded styling.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Message struct {`

Logic:
- Message renders one user turn.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(value string, opts Options) *Message`

Logic:
- New creates a user message component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (m *Message) SetText(value string)`

Logic:
- SetText replaces message text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (m *Message) Text() string`

Logic:
- Text returns raw message text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (m *Message) Invalidate() {}`

Logic:
- Invalidate exists for the component interface.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
