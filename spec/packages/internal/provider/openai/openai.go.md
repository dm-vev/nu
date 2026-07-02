# `internal/provider/openai/openai.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Phase 3 OpenAI adapter covers Chat Completions and Responses request/stream shapes.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement OpenAI Chat Completions and Responses provider adapters with no SDK
dependency.

## Code Style

Stdlib HTTP/JSON/SSE only. Keep request builders testable without network calls.
Never include API keys in errors.

## Types

### `type Config`

Logic:

- Hold base URL, API key, API kind, and optional HTTP client.

Acceptance:

- tests can use `httptest.Server`.

### `type Client`

Logic:

- Implement `provider.Streamer`.

Acceptance:

- validates shared provider requests before HTTP work.

## Functions

### `New(cfg Config) *Client`

Logic:

- Default base URL to OpenAI v1.
- Default API kind to `responses`.
- Default HTTP client to `http.DefaultClient`.

Acceptance:

- usable with only an API key for real opt-in smoke tests.

### `BuildChatPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert provider messages into Chat Completions messages.
- Include `model`, `stream: true`, and stream usage options.
- Represent tool results as `role: tool` with `tool_call_id`.

Acceptance:

- matches Chat Completions request shape used by tests.

### `BuildResponsesPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert provider messages into Responses `input` items.
- Include `model` and `stream: true`.

Acceptance:

- matches Responses request shape used by tests.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to the configured OpenAI endpoint.
- Add bearer auth and JSON headers.
- Emit provider `start`.
- Parse Chat or Responses SSE events into provider-neutral text/tool/done/error
  events.

Acceptance:

- Chat request shape is test-covered;
- Responses function-call streaming becomes tool call start/delta/end and
  terminal `tool_use`.

Tests:

- `TestNUF030OpenAIChatRequestShape`
- `TestNUF030OpenAIResponsesToolCallStream`

