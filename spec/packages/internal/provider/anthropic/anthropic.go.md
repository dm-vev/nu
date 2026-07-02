# `internal/provider/anthropic/anthropic.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Phase 3 Anthropic Messages adapter covers request shape and SSE text/tool parsing.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement Anthropic Messages provider adapter with no SDK dependency.

## Code Style

Stdlib HTTP/JSON/SSE only. Keep the request builder independent from network
transport.

## Functions

### `BuildMessagesPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user/assistant text messages into Anthropic `messages`.
- Convert tool results into user-role `tool_result` content blocks.
- Include `model`, `max_tokens`, and `stream: true`.

Acceptance:

- matches Messages request shape used by tests.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to `/v1/messages`.
- Add `x-api-key`, `anthropic-version`, and JSON headers.
- Parse Anthropic SSE `content_block_*`, `message_delta`, `message_stop`, and
  `error` events.

Acceptance:

- request shape is test-covered;
- text deltas and tool-use deltas normalize to provider events.

Tests:

- `TestNUF030AnthropicMessagesRequestShape`

