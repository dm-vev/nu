# `internal/provider/anthropic/messages.go`

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

Anthropic Messages API adapter.

## Code Style

Request conversion is explicit and golden-tested. Cache and thinking fields stay
local to this adapter.

## Functions

### `NewClient(cfg Config) *Client`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Configure endpoint, auth, version headers, beta headers, and HTTP client.

Acceptance:

- configures endpoint, auth, version headers, beta headers, and HTTP client.

### `(*Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Resolve API-key or subscription OAuth credential and required version headers.
- Convert system prompt and Nu messages to Anthropic `system` and `messages`.
- Convert Nu tool results to Anthropic `tool_result` blocks.
- Convert Nu tools to Anthropic tool schemas.
- Map thinking level/cache settings to Anthropic request fields.
- Create HTTP POST to `/v1/messages` with streaming enabled.
- Decode Anthropic event stream and map content block start/delta/stop, thinking
  deltas, tool use blocks, message delta usage, message stop, and error events
  to `provider.Event`.
- Classify overloaded/rate-limit/quota/auth errors for retry layer.

Acceptance:

- converts content blocks, thinking, tool use, tool results, usage, and errors;
- supports subscription and API-key auth.
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestNUF030AnthropicMessagesRequestShape`
