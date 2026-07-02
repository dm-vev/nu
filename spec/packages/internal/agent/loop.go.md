# `internal/agent/loop.go`

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

Provider/tool turn loop.
Consumes `spec/protocols/provider-stream.md`.

## Code Style

Keep the loop readable. Do not hide state transitions behind clever channels.

## Functions

### `runTurn(ctx context.Context, state *State, input TurnInput) error`

Logic:

- Build provider request from active session branch, system prompt, compaction
  summaries, queued steering, current model, and active tool definitions.
- Emit `turn_start` before calling provider.
- Start provider stream with the same `ctx` used for abort.
- For every provider event, call `handleProviderEvent` and emit corresponding
  message/update events.
- On normalized provider `error`, classify it and return to retry layer.
- On `done` with `stop`, `length`, or `content_filter`, persist assistant
  message and emit `turn_end`.
- On `done` with `tool_use`, verify every tool call is finalized, execute tools,
  persist tool results, append them to context, then start the next provider
  request in the same turn.
- Before each new provider request, drain steering queue according to settings.
- After final assistant stop and no steering remains, deliver follow-up queue if
  configured.
- Stop on context cancellation, abort, or unrecoverable validation error.

Acceptance:

- sends context to provider;
- accumulates assistant stream into messages;
- detects tool calls;
- feeds tool results back to provider;
- stops only on final assistant stop, abort, or unrecoverable error.

### `handleProviderEvent(state *State, ev provider.Event) error`

Logic:

- Switch only on normalized provider event type.
- Create assistant message on `start` if one is not active.
- Append `text_delta` and `thinking_delta` to content blocks by event index.
- Start, append, and finalize tool-call argument buffers by tool-call index.
- Merge usage events into assistant usage.
- Convert `done` into assistant stop reason.
- Convert `error` into typed provider error without mutating finalized messages.
- Return an error for impossible ordering, such as delta before start or done
  with unfinalized tool call.

Acceptance:

- converts provider events to message updates without provider-specific logic.

Tests:

- `TestNUF050ToolCallFeedsResultBackToProvider`
