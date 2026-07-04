# `internal/agent/loop.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Provider loop handles text, thinking, finalized tool calls, and rate-limit retry, preserves assistant tool-call history before tool results, emits tool args/results for UI/RPC, and rejects malformed tool-call ordering; queues remain future specs.

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
- Start accumulated messages from `TurnInput.History` and append the current user prompt.
- Include provider-facing tool definitions on every provider request.
- Emit `turn_start` before calling provider.
- Start provider stream with the same `ctx` used for abort.
- For every provider event, call `handleProviderEvent` and emit corresponding
  message/update events, including thinking deltas.
- On normalized provider `error`, return `provider.ErrStream` with context.
- On provider rate-limit setup or stream errors, emit `rate_limit`, wait a short backoff, and retry up to five times.
- On `done` with `tool_use`, execute finalized tool calls, append assistant
  tool-call messages followed by tool result messages, and start the next
  provider request in the same turn.
- On final `done`, append non-empty assistant text to accumulated messages and emit final `turn_end` with accumulated text.
- Stop on context cancellation, abort, premature stream close, or validation error.

Acceptance:

- sends context to provider;
- accumulates assistant stream into messages;
- preserves prior successful user/assistant turns in the next provider request;
- feeds assistant tool-call history and tool results back to provider;
- forwards tool definitions so models can actually call tools;
- stops on final provider done, abort, provider error, or unrecoverable error.
- retries rate-limit failures up to five times before returning the final error.

### `runProviderStream(ctx context.Context, state *State) (string, error)`

Logic:

- Call one provider stream attempt.
- Retry only errors wrapping `provider.ErrRateLimit`.
- Emit `rate_limit` with `attempt` and `max` before each retry.
- Use a short fixed backoff that respects context cancellation.

Acceptance:

- `TestAgentRetriesRateLimitBeforeFailing` fails if rate limits are not retried.

### `runProviderStreamOnce(ctx context.Context, state *State) (string, error)`

Logic:

- Build and validate one provider request.
- Drain provider events until done/error/premature close.
- Convert stream `ErrorClass=rate_limit` into `provider.ErrRateLimit`.

Acceptance:

- Non-rate provider errors still return `provider.ErrStream`.

### `handleProviderEvent(state *State, ev provider.Event) error`

Logic:

- Switch only on normalized provider event type.
- Emit `message_start` on provider start.
- Append `text_delta` to the accumulated response text and emit a live update.
- Emit `thinking_delta` as structured `message_update` data without adding it
  to final assistant text.
- Assemble tool-call arguments from start/delta/end events by index.
- Reject missing tool id/name, duplicate end, and deltas after end.
- Emit `message_end` on provider done.
- Leave provider error handling to `runTurn`.
- Return an error for unknown normalized events.

Acceptance:

- converts start/text/thinking/tool/done provider events without provider-specific logic.

### `executeToolCalls(ctx context.Context, state *State) ([]provider.Message, error)`

Logic:

- Require at least one finalized tool call after a `tool_use` stop.
- Sort pending tool calls by provider index for deterministic continuation.
- Reject unfinished tool calls and missing tool implementations.
- Append the assistant tool-call message before executing and appending the
  tool result message.
- Emit tool start/end events around each tool execution, including raw
  arguments, result content, and error state for TUI/RPC rendering.
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
- `TestNUF050ThinkingDeltaEmitsStructuredMessageUpdate`
- `TestAgentRetriesRateLimitBeforeFailing`
- `TestAgentStopsAfterFiveRateLimitRetries`
