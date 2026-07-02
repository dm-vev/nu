# `internal/message/event.go`

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

Define runtime event stream types.

## Code Style

Events are immutable values after emission. JSON shape is part of public API.

## Types

### `Event`

Logic:

- Represent agent_start/agent_end, turn_start/turn_end, message_start/update/end,
  tool_execution_start/update/end, queue_update, compaction_start/end,
  auto_retry_start/end, session, and protocol error events.
- Keep event payloads immutable after emission.
- Keep JSON field names stable because JSON/RPC/export consumers depend on
  them.
- Store protocol-specific fields only when the protocol spec requires them.

Acceptance:

- covers agent, turn, message, tool execution, queue, compaction, retry, and
  session events;
- marshals as one JSON object per event.

## Functions

### `MarshalJSONL(event Event) ([]byte, error)`

Logic:

- Validate event type and payload type match.
- Marshal compact JSON.
- Do not append LF; JSONL writers own delimiters.
- Return an error for unknown event types or payloads that cannot be represented
  in JSON.

Acceptance:

- appends no extra whitespace except caller-managed newline;
- rejects unknown event payloads.

Tests:

- `TestNUF061JSONModeEmitsSessionThenEvents`
