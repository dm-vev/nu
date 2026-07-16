# `internal/tui/message_mutate.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Streaming mutations preserve part ordering.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Text/thinking/tool mutation paths are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Mutate structured messages as agent events arrive.

## Code Style

Keep mutations explicit. Do not hide ordering behavior behind reflection or maps.

## Acceptance Criteria

- Empty deltas are ignored.
- Adjacent text deltas coalesce with text only.
- Adjacent thinking deltas coalesce with thinking only.
- Tool finalization does not disturb text or thinking parts.

## Functions

### `func (m *Message) AppendText(delta string)`

Logic:
- Ignore empty deltas.
- Append to the last text part if it is adjacent.
- Otherwise append a new text part.

Acceptance:
- Streaming assistant text remains ordered around thinking/tool parts.

### `func (m *Message) AppendThinking(delta string)`

Logic:
- Ignore empty deltas.
- Append to the last thinking part if it is adjacent.
- Otherwise append a new thinking part.

Acceptance:
- Model reasoning is visible as its own gray/italic part.

### `func (m *Message) ReplaceText(value string)`

Logic:
- Replace the latest text part when it exists.
- Append a text part when no text part exists.

Acceptance:
- Final text updates do not delete tool or thinking parts.

### `func (m *Message) AddTool(id string, name string, arguments string)`

Logic:
- Append a pending tool part with raw id, name, and arguments.

Acceptance:
- TUI can render a pending command block immediately on `tool_start`.

### `func (m *Message) FinishTool(id string, result string, failed bool)`

Logic:
- Walk backward to find the matching tool part.
- Store the result and set success/error state.
- Leave state unchanged when no matching tool is found.

Acceptance:
- TUI updates the latest matching block on `tool_end`.
