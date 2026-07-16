# Internal SDK Package Index

Status: **IMPLEMENTED: the imported baseline uses the approved balanced hierarchy**.

The SDK fork lives directly under `internal/`. This file groups SDK ownership
without limiting features or promising to preserve upstream import paths. Every
SDK feature and API behavior currently imported from the pinned baseline remains
available. The target below is exhaustive and source-breaking. No feature or API
behavior is deleted, and no compatibility wrapper preserves an old import path.

```text
app/{auth,cli}
agent/{config,plans,guardrails,prompts}
llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
tools/{coding,search,image,graphrag}
data/{embedding,weaviate/{graph,vector},sql,storage}
task/{service,workflow,orchestration}
telemetry/{otel,langfuse}
transport/{grpc/pb,http,a2a,ui}
tui/{core,editor,engine,input,message,terminal,components}
standalone: agentui config contracts memory multitenancy mcp model rpc session testkit
```

| Imported feature family | Approved owner | Requirement |
|---|---|---|
| Agent construction, `AgentContext`, generation, streaming, and sub-agent orchestration | `internal/agent` | `NUF-200`, `NUF-201`, `NUF-212` |
| Deployment config/loading, execution plans, guardrails, and prompt templates/stores | `internal/agent/{config,plans,guardrails,prompts}` | `NUF-200`, `NUF-212` |
| Shared SDK interfaces and event/value contracts | `internal/contracts` | `NUF-200`, `NUF-212` |
| SDK configuration not specific to one agent | `internal/config` | `NUF-200`, `NUF-212` |
| Shared LLM types, retry, and structured-output orchestration | `internal/llm` | `NUF-201`, `NUF-203`, `NUF-212` |
| Providers and provider streaming | `internal/llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}` | `NUF-203`, `NUF-212` |
| Conversation memory | `internal/memory` | `NUF-210`, `NUF-212` |
| Tenancy | `internal/multitenancy` | `NUF-212` |
| MCP client/server, transport, sampling, and management | `internal/mcp` | `NUF-211`, `NUF-212` |
| Shared tool types, registry, and execution orchestration | `internal/tools` | `NUF-204`, `NUF-211`, `NUF-212` |
| Coding, search, image, and GraphRAG tools | `internal/tools/{coding,search,image,graphrag}` | `NUF-204`, `NUF-212` |
| Data package index, with no forwarding API | `internal/data` | `NUF-210`, `NUF-212` |
| Embedders and generic metadata filtering/evaluation | `internal/data/embedding` | `NUF-210`, `NUF-212` |
| GraphRAG Weaviate store | `internal/data/weaviate/graph` | `NUF-210`, `NUF-212` |
| Vector Weaviate store | `internal/data/weaviate/vector` | `NUF-210`, `NUF-212` |
| PostgreSQL and Supabase adapters | `internal/data/sql` | `NUF-210`, `NUF-212` |
| Storage contract plus local and GCS implementations | `internal/data/storage` | `NUF-210`, `NUF-212` |
| Canonical/core and legacy task models, executors, planners, and shared contracts/options | `internal/task` | `NUF-212` |
| In-memory/core task services, API support, adapters, and compatibility conversion | `internal/task/service` | `NUF-212` |
| Workflow models and execution | `internal/task/workflow` | `NUF-212` |
| LLM/code/workflow orchestrators, handoffs, registries, and routers | `internal/task/orchestration` | `NUF-212` |
| Shared telemetry contracts and fan-out | `internal/telemetry` | `NUF-201`, `NUF-212` |
| OpenTelemetry/logging and Langfuse | `internal/telemetry/{otel,langfuse}` | `NUF-201`, `NUF-212` |
| Shared transport types/construction | `internal/transport` | `NUF-212` |
| gRPC, HTTP/microservice, A2A, and UI transports | `internal/transport/{grpc,http,a2a,ui}` | `NUF-212` |
| Generated protobuf only | `internal/transport/grpc/pb` | `NUF-212` |

Nu-owned `internal/app`, `internal/app/auth`, `internal/app/cli`,
`internal/agentui`, `internal/model`, `internal/session`, `internal/rpc`,
`internal/tui`, every listed `internal/tui/*` package, `internal/tools/coding`,
and `internal/testkit` are integration consumers/additions, not imported SDK
families. `cmd/nu` is the only command package. SDK-owned code must not import a
Nu-owned package.

For dependency-direction checks, every owner in the table is SDK-owned except
the documented Nu `tools/coding` addition. The Nu packages named above are
Nu-owned. No other production Go package below `internal/` and no production Go
package at the `internal/` root is part of the approved target. The complete
imported source and owning tests, rather than table granularity, define the
feature baseline.

Concrete remote clients belong to their `internal/transport/*` family, while
`internal/agent` consumes only transport-neutral contracts supplied at
composition. Generated protobuf belongs to `internal/transport/grpc/pb`; no
other package owns generated protobuf.

## Package And File Rules

- A package is a cohesive domain or dependency boundary. Keep production
  packages within this target and move all imported behavior and tests with an
  ownership change.
- Do not create compatibility wrappers for moved or merged package paths.
- Do not create a subpackage for one helper. Every listed subpackage must own a
  cohesive feature family; merge a stray helper into its caller.
- Root packages own shared types and cross-subpackage orchestration only.
- Use ordinary filenames inside subpackages (`client.go`, not
  `openai_client.go`). Package paths already supply the domain prefix.
- Every reusable TUI component belongs to the single `internal/tui/components`
  package; component-per-package trees are forbidden.
- A production file has one cohesive responsibility. Behavior-neutral splits
  within a package are allowed.
- A non-generated production `.go` file over 300 lines must be split or have a
  documented exception in `../backend.md`; generated and test files are exempt.
- Generated protobuf under `internal/transport/grpc/pb` is regenerated from its
  `.proto` source with the generator command and pinned `protoc`/plugin versions
  recorded in `../backend.md`. Generated Go and descriptor bytes are never
  hand-edited.

Provenance, local patches, and update procedure are canonical in
`../backend.md`.
