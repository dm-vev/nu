# `internal/provider/anthropic/anthropic.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Phase 3 Anthropic Messages adapter covers request shape, assistant tool-use history, tool results, and SSE text/thinking/tool parsing.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement Anthropic Messages provider adapter with no SDK dependency.

## Code Style

Stdlib HTTP/JSON/SSE only. Keep the request builder independent from network
transport.

## Functions

### `BuildMessagesPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user/assistant text messages into Anthropic `messages`.
- Convert assistant tool-call history into assistant-role `tool_use` content blocks.
- Convert tool results into user-role `tool_result` content blocks.
- Include `model`, `max_tokens`, and `stream: true`.

Acceptance:

- matches Messages request shape used by tests;
- preserves assistant tool-use history before tool results.

### `messagesMessage(message provider.Message) map[string]any`

Logic:

- Convert assistant tool-call history into assistant `tool_use` content.
- Convert tool results into user-role `tool_result` content.
- Preserve user/assistant text messages as Anthropic role/content pairs.
- Normalize unknown roles to user text messages.

Acceptance:

- payload messages preserve the provider-required tool_use/tool_result order.

### `decodeJSONOrText(raw string) any`

Logic:

- Decode JSON arguments/results when possible.
- Fall back to `{ "text": raw }` when the value is not valid JSON.

Acceptance:

- malformed tool arguments do not panic or fail payload construction.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to `/v1/messages`.
- Add `x-api-key`, `anthropic-version`, and JSON headers.
- Parse Anthropic SSE `content_block_*`, `message_delta`, `message_stop`, and
  `error` events.
- Preserve `thinking_delta` blocks as provider `EventThinking`.
- Return `provider.ErrRateLimit` for HTTP 429 before stream setup.
- Normalize Anthropic `rate_limit_error` and `overloaded_error` stream errors
  to `ErrorClass=rate_limit`.

Acceptance:

- request shape is test-covered;
- text, thinking, and tool-use deltas normalize to provider events.
- rate-limit failures can be retried by `internal/agent`.

Tests:

- `TestNUF030AnthropicMessagesRequestShape`
- `TestAnthropicPayloadIncludesAssistantToolUse`
- `TestAnthropicThinkingDeltaIsPreserved`
