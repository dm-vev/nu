# `internal/tui/events.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Routes agent events into status and structured messages.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Text/thinking/tool event paths are covered by TUI tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Translate agent events into TUI message/status state and trigger exactly one
render per event.

## Acceptance Criteria

- `message_update` with `thinking_delta` appends a thinking part.
- `message_update` with text delta appends a text part.
- `tool_start` appends a pending tool part with arguments.
- `tool_end` finalizes a tool part with result and error state.
- `rate_limit` sets the footer notice and alert retry status.

## Functions

### `func (a *App) Emit(ev agent.Event)`

Logic:
- Lock TUI state, apply event-specific message/status changes, rebuild chat,
  unlock, then render.
- Clear stale footer notices once normal turn/message/end events resume.

Acceptance:
- Each event mutates state once and renders once.

### `func (a *App) setFooterNoticeLocked(value string)`

Logic:
- Update footer notice through footer options while caller owns the app lock.

Acceptance:
- Rate-limit status can appear on the path/footer line without mutating unrelated footer fields.

### `func eventText(data any, key string) string`

Logic:
- Extract a string field from `map[string]string` or `map[string]any`.

Acceptance:
- TUI can consume events from direct Go tests and JSON/RPC decoded payloads.

### `func eventBool(data any, key string) bool`

Logic:
- Extract bool fields from `map[string]any` or string bool fields from
  `map[string]string`.

Acceptance:
- Tool error state survives both direct events and decoded events.
