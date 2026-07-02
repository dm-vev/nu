# `internal/provider/stream.go`

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

Provider-neutral streaming event types and helpers.
Implements `spec/protocols/provider-stream.md`.

## Code Style

Events should be small and ordered. Stream parsing errors include provider name
and partial context without leaking secrets.

## Types

### `Event`

Logic:

- Represent only the normalized events from `spec/protocols/provider-stream.md`.
- Keep vendor event names out of this type.
- Store tool call argument deltas as raw bytes or strings until `tool_call_end`.
- Include enough metadata on `error` for retry classification without secrets.

Acceptance:

- covers start, text delta, thinking delta, tool call delta/end, usage, done,
  and error.
- matches event ordering and terminal-event rules in
  `spec/protocols/provider-stream.md`.

## Functions

### `Collect(ch <-chan Event) ([]Event, error)`

Logic:

- Initialize an empty result slice and terminal-state flag.
- Range over the event channel until it closes or a terminal event arrives.
- Append non-terminal events in arrival order.
- On `done`, append it, mark terminal success, and stop draining.
- On `error`, return the events seen so far plus a typed provider error.
- If the channel closes without `done` or `error`, return an unexpected EOF
  provider error unless the stream was cancelled.
- Reject any event received after a terminal event in tests through the stream
  validator helper.

Acceptance:

- drains until done or error;
- respects caller cancellation through channel closure behavior.
- detects streams that close before a terminal event.

Tests:

- `TestProviderCollectStopsOnError`
