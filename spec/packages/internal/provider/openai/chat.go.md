# `internal/provider/openai/chat.go`

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

OpenAI Chat Completions adapter.

## Code Style

All HTTP requests use injected `*http.Client`. Request JSON has golden tests.

## Functions

### `NewChatClient(cfg Config) *ChatClient`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Store base URL, auth resolver, headers, compat flags, and HTTP client.

Acceptance:

- stores base URL, auth resolver, headers, compat flags, and HTTP client.

### `(*ChatClient) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Resolve auth key and headers immediately before request execution.
- Convert Nu messages to Chat Completions messages, using developer role only
  when compat flags allow it.
- Convert Nu tools to OpenAI `tools[].function` schema and force object
  arguments.
- Create HTTP POST to `/chat/completions` with streaming enabled.
- On non-2xx, parse provider error body enough to classify without leaking
  secrets.
- Decode SSE frames with `sse.go`.
- Convert content deltas, reasoning-compatible fields when present, tool call
  deltas, usage, finish reason, and errors to `provider.Event`.
- Close response body on stream completion or cancellation.

Acceptance:

- emits provider-neutral events from SSE;
- maps tools to Chat Completions tool schema;
- respects cancellation and retry-relevant HTTP status codes.
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestNUF030OpenAIChatRequestShape`
