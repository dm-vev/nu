# SDK Provider Stream Boundary

The former Nu provider-neutral stream is removed. Provider streaming is owned by
`internal/contracts.StreamEvent`, `StreamingLLM`, the shared `internal/llm`
orchestration, and provider implementations under `internal/llm/*`.

SDK stream event types are message start, content delta/complete, message stop,
error, tool use/result, and thinking. Nu application code must consume only
`AgentStreamEvent` through `internal/agent.Agent.RunStream`; it cannot import a
provider client to implement a second tool loop.

Provider conformance is proved by each owning `internal/llm/*` package. Nu adds
only app-level mocked HTTP tests for auth/model/base-URL composition and
protocol-stdout isolation.
