# `internal/tui/components/assistantmessage/message.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Stores structured assistant parts for rendering.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Structured parts are covered by component tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Expose assistant messages as renderable components while preserving the legacy
single-string constructor for narrow tests.

## Code Style

Keep state copying explicit. Do not let component callers mutate internal part
slices after construction.

## Acceptance Criteria

- `New` creates a text-only assistant message.
- `NewMessage` copies structured parts from `internal/tui/message`.
- `Text` returns concatenated visible text parts only.

## Types

### `type Message struct`

Logic:
- Store ordered assistant parts and normalized render options.

Acceptance:
- Render can distinguish text, thinking, and tool parts.

## Functions

### `func New(value string, opts Options) *Message`

Logic:
- Create a text-only assistant component for compatibility with simple tests.

Acceptance:
- Existing text rendering tests keep using this constructor.

### `func NewMessage(value tuimessage.Message, opts Options) *Message`

Logic:
- Copy structured assistant parts and normalize options.

Acceptance:
- TUI chat rebuild uses this constructor for agent event output.

### `func (m *Message) SetText(value string)`

Logic:
- Replace all current parts with one text part.

Acceptance:
- Legacy text update behavior remains deterministic.

### `func (m *Message) Text() string`

Logic:
- Concatenate visible text parts and ignore thinking/tool parts.

Acceptance:
- Tests can inspect human-visible assistant text without tool metadata.

### `func (m *Message) Invalidate()`

Logic:
- Satisfy the invalidatable component convention.

Acceptance:
- Container invalidation can call it safely.
