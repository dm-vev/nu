# `internal/provider/openai/openai.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Phase 3 OpenAI adapter covers Chat Completions and Responses request/stream shapes, assistant tool-call history, and deterministic tool-end ordering.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

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
- Include `tools` and `tool_choice: auto` when tool definitions are present.
- Represent assistant tool-call history as Chat Completions `tool_calls`.
- Represent tool results as `role: tool` with `tool_call_id`.

Acceptance:

- matches Chat Completions request shape used by tests;
- advertises function tools for OpenAI-compatible providers such as Fireworks;
- preserves assistant tool-call history before tool results.

### `chatTools(tools []provider.ToolDefinition) []map[string]any`

Logic:

- Convert provider-neutral tool definitions into Chat Completions function tools.

Acceptance:

- Each tool has `type: function`, name, description, and parameters.

### `chatMessage(message provider.Message) map[string]any`

Logic:

- Convert assistant tool-call history into one Chat Completions assistant
  message with `tool_calls`.
- Convert tool results into `role: tool` messages with `tool_call_id`.
- Preserve simple user/assistant text messages as role/content pairs.

Acceptance:

- Chat payloads include the assistant tool-call entry required before tool
  result messages.

### `BuildResponsesPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert provider messages into Responses `input` items.
- Include `tools` when tool definitions are present.
- Represent assistant tool-call history as `function_call` input items.
- Represent tool results as `function_call_output` input items.
- Include `model` and `stream: true`.

Acceptance:

- matches Responses request shape used by tests;
- preserves function-call history before function-call output.

### `responsesTools(tools []provider.ToolDefinition) []map[string]any`

Logic:

- Convert provider-neutral tool definitions into Responses function tools.

Acceptance:

- Each tool has `type: function`, name, description, and parameters.

### `responsesInput(message provider.Message) map[string]any`

Logic:

- Convert assistant tool-call history into Responses `function_call` input.
- Convert tool results into `function_call_output` input.
- Preserve simple user/assistant messages as role/content items.

Acceptance:

- Responses payloads include function-call history before tool output.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to the configured OpenAI endpoint.
- Add bearer auth and JSON headers.
- Emit provider `start`.
- Parse Chat or Responses SSE events into provider-neutral text/tool/done/error
  events.
- Return `provider.ErrRateLimit` for HTTP 429 before stream setup.
- Normalize streamed Responses errors containing rate-limit data to
  `ErrorClass=rate_limit`.

Acceptance:

- Chat request shape is test-covered;
- Responses function-call streaming becomes tool call start/delta/end and
  terminal `tool_use`.
- rate-limit failures can be retried by `internal/agent`.

Tests:

- `TestNUF030OpenAIChatRequestShape`
- `TestNUF030OpenAIResponsesToolCallStream`
- `TestOpenAIChatPayloadIncludesAssistantToolCalls`
- `TestOpenAIChatPayloadIncludesToolDefinitions`
- `TestOpenAIResponsesPayloadIncludesFunctionCallHistory`
- `TestOpenAIResponsesPayloadIncludesToolDefinitions`
