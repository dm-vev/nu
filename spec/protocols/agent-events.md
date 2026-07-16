# Agent Event Adaptation

The backend event contract is `contracts.AgentStreamEvent`:

- `content`
- `thinking`
- `tool_call`
- `tool_result`
- `error`
- `complete`

`internal/agentui` maps it to existing Nu application events:

| SDK | Nu event |
|---|---|
| prompt accepted | `turn_start`, `message_start` |
| non-empty `content` | `message_update {delta}` |
| `thinking` | `message_update {kind:thinking, thinking_delta}` |
| `tool_call` | `tool_call_start`, `tool_call_end`, `tool_start` |
| `tool_result` | `tool_end` |
| `error` | prompt error; no successful `turn_end` |
| clean close/complete | `message_end`, `turn_end {text}` |

Empty SDK content lifecycle events and metadata-only iteration boundaries do not
become visible text. Tool IDs, names, raw arguments, results, and status are
preserved. TUI/RPC tests own this compatibility shape; SDK package tests own the
backend stream itself.
