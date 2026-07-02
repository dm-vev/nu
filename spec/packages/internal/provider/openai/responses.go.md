# `internal/provider/openai/responses.go`

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

OpenAI Responses API adapter.

## Code Style

Keep Responses-specific event names isolated here. Agent sees only neutral
provider events.

## Functions

### `NewResponsesClient(cfg Config) *ResponsesClient`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Configure base URL, auth, headers, and HTTP client.

Acceptance:

- configures base URL, auth, headers, and HTTP client.

### `(*ResponsesClient) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Resolve auth key and headers immediately before request execution.
- Convert Nu messages into Responses `input` items, preserving tool result
  linkage.
- Convert text and image content blocks according to model capabilities.
- Convert tools to Responses function tool definitions.
- Map thinking level to Responses reasoning fields when configured.
- Create HTTP POST to `/responses` with stream enabled.
- Decode SSE event names locally and map output text, reasoning, function-call
  argument deltas, usage, completed, incomplete, and error events to
  `provider.Event`.
- Close response body on stream completion or cancellation.

Acceptance:

- converts response text, reasoning, function call, usage, and done events;
- supports image input where model allows it.
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestNUF030OpenAIResponsesToolCallStream`
