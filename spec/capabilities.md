# Backend Capability Matrix

This matrix records which backend capabilities Nu currently exposes. It does not
limit the complete imported SDK feature set required by `NUF-212`.

| Capability | Backend owner | Nu CLI status |
|---|---|---|
| Agent Run/RunDetailed/RunStream | `internal/agent` | connected |
| OpenAI-compatible streaming | `internal/llm/openai` | connected |
| Anthropic | `internal/llm/anthropic` | connected |
| Gemini | `internal/llm/gemini` | connected |
| Claude on Bedrock | `internal/llm/anthropic` | connected |
| Nu coding tools | `internal/tools/coding` -> `contracts.Tool` | connected |
| Bounded conversation memory | `internal/memory` | connected |
| MCP client tools/resources/prompts over required transports | `internal/mcp` | specified; configuration exposure pending |
| Retry | `internal/llm` | connected through SDK |
| Diagnostics-safe logging and tracing | `internal/telemetry/otel` | connected |
| Branchable Nu sessions | `internal/session` | Nu-owned, partial app integration |
| TUI | `internal/tui/...` via `internal/agentui` | connected |

## Exposure Rule

A capability can be exposed through Nu only after its functional, trust, secret,
protocol-stdout, and integration-test requirements are specified. Lack of Nu
exposure is not permission to remove its imported SDK behavior or owning tests.
