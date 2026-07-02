# `internal/agent/loop.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Provider loop handles text and finalized tool calls, preserves assistant tool-call history before tool results, and rejects malformed tool-call ordering; retry/queues remain future specs.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Provider/tool turn loop for the Phase 1 headless spine. Consumes
`spec/protocols/provider-stream.md`.

## Code Style

Keep the loop readable. Do not hide state transitions behind clever channels.

## Functions

### `runTurn(ctx context.Context, state *State, input TurnInput) error`

Logic:

- Build a provider request from current provider/model fields and accumulated messages.
- Emit `turn_start` before calling provider.
- Start provider stream with the same `ctx` used for abort.
- For every provider event, call `handleProviderEvent` and emit corresponding
  message/update events.
- On normalized provider `error`, return `provider.ErrStream` with context.
- On `done` with `tool_use`, execute finalized tool calls, append assistant
  tool-call messages followed by tool result messages, and start the next
  provider request in the same turn.
- On final `done`, emit final `turn_end` with accumulated text.
- Stop on context cancellation, abort, premature stream close, or validation error.

Acceptance:

- sends context to provider;
- accumulates assistant stream into messages;
- feeds assistant tool-call history and tool results back to provider;
- stops on final provider done, abort, provider error, or unrecoverable error.

### `handleProviderEvent(state *State, ev provider.Event) error`

Logic:

- Switch only on normalized provider event type.
- Emit `message_start` on provider start.
- Append `text_delta` to the accumulated response text and emit a live update.
- Assemble tool-call arguments from start/delta/end events by index.
- Reject missing tool id/name, duplicate end, and deltas after end.
- Emit `message_end` on provider done.
- Leave provider error handling to `runTurn`.
- Return an error for unknown normalized events.

Acceptance:

- converts start/text/tool/done provider events without provider-specific logic.

### `executeToolCalls(ctx context.Context, state *State) ([]provider.Message, error)`

Logic:

- Require at least one finalized tool call after a `tool_use` stop.
- Sort pending tool calls by provider index for deterministic continuation.
- Reject unfinished tool calls and missing tool implementations.
- Append the assistant tool-call message before executing and appending the
  tool result message.
- Emit tool start/end events around each tool execution.
- Wrap context cancellation and tool execution errors with tool name context.

Acceptance:

- returned provider messages preserve assistant tool-call history and tool
  result order;
- missing or unfinished tool calls fail before the next provider request.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050AbortStopsProviderAndTools`
- `TestNUF050RejectsMalformedToolCallStream`
