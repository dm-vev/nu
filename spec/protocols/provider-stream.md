# Provider Stream Protocol

## Purpose

Normalize vendor streams into one event model consumed by `internal/agent`.
Provider adapters own vendor quirks; the agent must never branch on OpenAI,
Anthropic, Gemini, Bedrock, or compatibility-server event names.

## Event Ordering

Every provider stream emits:

1. `start`
2. zero or more content/tool/usage events
3. exactly one terminal event: `done` or `error`

No events are emitted after `done` or `error`. Cancellation may close the stream
without `done`; the adapter must then return/emit a cancellation error visible
to the caller.

## Event Types

### `start`

Fields:

- `provider`
- `api`
- `model`
- `request_id`, optional

Rules:

- first event only;
- no token usage here.

### `text_delta`

Fields:

- `index`
- `delta`

Rules:

- appends to assistant visible text block at `index`;
- empty deltas are ignored unless vendor uses them as keepalive, in which case
  they are not forwarded.

### `thinking_delta`

Fields:

- `index`
- `delta`
- `signature`, optional

Rules:

- appends to reasoning/thinking block at `index`;
- adapters must not expose provider-private encrypted thinking blobs unless the
  provider requires round-trip preservation.

### `tool_call_start`

Fields:

- `index`
- `id`
- `name`

Rules:

- begins a tool call content block;
- `id` must be stable and non-empty; if vendor omits it, adapter generates a
  deterministic id scoped to the assistant message.

### `tool_call_delta`

Fields:

- `index`
- `delta`

Rules:

- appends raw JSON argument bytes for the tool call at `index`;
- deltas can split UTF-8 and JSON tokens; adapter or assembler must preserve
  bytes until final decode.

### `tool_call_end`

Fields:

- `index`

Rules:

- finalizes the tool call;
- accumulated arguments must decode to a JSON object;
- malformed JSON becomes provider stream `error`, not a tool error.

### `usage`

Fields:

- `input_tokens`
- `output_tokens`
- `cache_read_tokens`
- `cache_write_tokens`
- `cost`, optional

Rules:

- usage may arrive before or after `done`;
- later usage events replace missing fields or add deltas according to adapter
  contract; adapters must normalize final assistant usage before `done`.

### `done`

Fields:

- `stop_reason`: `stop`, `length`, `tool_use`, `content_filter`, `aborted`
- `usage`, optional final usage

Rules:

- terminal success event;
- `tool_use` means agent must execute finalized tool calls and continue.

### `error`

Fields:

- `class`: `transient`, `auth`, `quota`, `rate_limit`, `unsupported`,
  `cancelled`, `fatal`
- `message`
- `retry_after_ms`, optional

Rules:

- terminal failure event;
- message must not contain API keys, auth headers, or raw request body.

## Assembly Rules

- Assistant message content order is event `index` order.
- Text/thinking blocks with the same index append in arrival order.
- Tool call deltas may interleave across indexes.
- A tool call is executable only after `tool_call_end`.
- `done` with `tool_use` is invalid if any tool call is unfinalized.
- Duplicate `tool_call_start` for the same index is an error unless it repeats
  identical id/name before any delta; adapters should collapse that case.

## Tests

- interleaved tool-call delta assembly;
- malformed final tool JSON returns provider error;
- cancellation closes stream and classifies as cancelled;
- usage before and after content normalizes to one assistant usage;
- no event after terminal event is accepted.
