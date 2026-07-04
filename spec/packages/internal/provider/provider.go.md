# `internal/provider/provider.go`

## Status

Current: IMPLEMENTED
Implementation Commit: a94e00c
Implementation Comments: Streamer, request validation, normalized text/thinking/tool events, and stream collection live here while provider has one consumer.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define the provider adapter contract and current provider-neutral request/event
types consumed by `internal/agent`.

## Code Style

Keep the package flat while the agent is the only consumer. Split files only
when adapters add enough parsing code to justify it.

## Types

### `type Streamer interface`

Logic:

- Expose only `Stream(ctx, req)` so adapters stay replaceable in tests.
- Require adapters to honor context cancellation and close their event channel.
- Return setup errors before starting a stream when request/auth configuration is invalid.
- Emit provider-neutral events after streaming begins.
- Return `ErrRateLimit` or emit `EventError` with `ErrorClass=rate_limit` for retryable rate-limit failures.

Acceptance:

- has `Stream(ctx context.Context, req Request) (<-chan Event, error)`;
- respects context cancellation.

### `type Request`

Logic:

- Carry provider id, API kind, model id, and ordered prompt messages.
- Carry optional provider-neutral tool definitions for providers that support
  function/tool calling.
- Stay provider-neutral; adapters translate this into HTTP/provider payloads.
- Leave auth and transport settings outside this struct.

Acceptance:

- includes provider, API, model, messages, and optional tools.

### `type ToolDefinition`

Logic:

- Describe one callable tool with name, description, and JSON schema parameters.
- Stay provider-neutral so adapters can translate it into their native tool format.

Acceptance:

- OpenAI-compatible adapters can advertise built-in tools to models.

### `type Event`

Logic:

- Normalize provider stream chunks into start, text delta, thinking delta, tool
  call start, tool call delta, tool call end, done, or error.
- Carry only fields the agent can consume now.
- Keep provider-specific payloads inside adapters until a feature needs them.
- Carry `ErrorClass` and optional retry delay metadata for retryable provider errors.

Acceptance:

- represents start, text delta, thinking delta, tool call, done, and error events.
- represents rate-limit errors without provider-specific branching in the TUI.

## Functions

### `ValidateRequest(req Request) error`

Logic:

- Reject missing provider, API, model, or message list before network work.
- Return `ErrInvalidRequest` wrapped with field context.
- Do not validate provider-specific auth here.

Acceptance:

- rejects incomplete requests.

### `Collect(ch <-chan Event) ([]Event, error)`

Logic:

- Drain events in order until `EventDone`.
- Return collected events and nil on done.
- Return `ErrStream` on provider error or premature channel close.

Acceptance:

- preserves event order;
- returns `ErrStream` for error and EOF cases.

Tests:

- `TestProviderCollectStopsAtDone`
- `TestProviderCollectRejectsErrorEvent`
- `TestProviderCollectRejectsUnexpectedEOF`
