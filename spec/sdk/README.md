# Public SDK Package Index

Status: **IMPLEMENTED: the imported baseline has a public agent SDK boundary**.

The public SDK lives at `github.com/dm-vev/nu/agent`,
`github.com/dm-vev/nu/contracts`, and `github.com/dm-vev/nu/telemetry`.
Implementation-only Nu integrations remain under `internal/`. This file groups
SDK ownership without limiting features or promising to preserve upstream import
paths. Every SDK feature and API behavior currently imported from the pinned
baseline remains available. The target below is exhaustive and source-breaking.
No feature or API behavior is deleted, and no compatibility wrapper preserves an
old import path.

```text
app/{auth,cli}
agent/{context,config,execution,generation,guardrails,image,mcp,plans,prompts,providers,remote,tools}
llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
tools/{agent,calculator,registry,coding,search,image/{edit,generation},graphrag}
memory/{conversation,history,redis,vector,factory}
mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}
data/{embedding/{gemini,openai},weaviate/{graph,vector},sql/{postgres,supabase},storage/{gcs,local}}
task/{service/{bridge},workflow,orchestration/{llm}}
telemetry/{otel,langfuse}
transport/{remote,grpc/{client,server,microservice,pb},http/server,a2a/{card,client,server,tool},ui/{server,trace}}
tui/{core,editor,engine,input,message,terminal,components}
public: agent/{context,config,execution,generation,guardrails,image,mcp,plans,prompts,providers,remote,tools} contracts telemetry/{otel,langfuse}
standalone: internal/agentui internal/config internal/multitenancy internal/model internal/rpc internal/session internal/testkit
```

| Imported feature family | Approved owner | Requirement |
|---|---|---|
| Public Agent implementation and orchestration | `github.com/dm-vev/nu/agent` | `NUF-200`, `NUF-201`, `NUF-212` |
| Agent context and sub-agent invocation values | `github.com/dm-vev/nu/agent/context` | `NUF-200`, `NUF-212` |
| Execution, MCP, providers, and remote lifecycle | `github.com/dm-vev/nu/agent/{execution,generation,image,mcp,plans,providers,remote,tools}` | `NUF-200`, `NUF-212` |
| Deployment config/loading, execution plans, guardrails, and prompt templates/stores | `github.com/dm-vev/nu/agent/{config,plans,guardrails,prompts}` | `NUF-200`, `NUF-212` |
| Shared SDK interfaces and event/value contracts | `github.com/dm-vev/nu/contracts` | `NUF-200`, `NUF-212` |
| SDK configuration not specific to one agent | `internal/config` | `NUF-200`, `NUF-212` |
| Shared LLM types, retry, and structured-output orchestration | `internal/llm` | `NUF-201`, `NUF-203`, `NUF-212` |
| Providers and provider streaming | `internal/llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}` | `NUF-203`, `NUF-212` |
| Conversation memory | `internal/memory/{conversation,history,redis,vector,factory}` | `NUF-210`, `NUF-212` |
| Tenancy | `internal/multitenancy` | `NUF-212` |
| MCP client/server, transport, sampling, and management | `internal/mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}` | `NUF-211`, `NUF-212` |
| Agent tools, Calculator, and registry | `internal/tools/{agent,calculator,registry}` | `NUF-204`, `NUF-211`, `NUF-212` |
| Coding, search, image, and GraphRAG tools | `internal/tools/{coding,search,image/{edit,generation},graphrag}` | `NUF-204`, `NUF-212` |
| Data package index, with no forwarding API | `internal/data` | `NUF-210`, `NUF-212` |
| Shared embedding contracts and metadata filtering/evaluation | `internal/data/embedding` | `NUF-210`, `NUF-212` |
| Gemini and OpenAI embedders | `internal/data/embedding/{gemini,openai}` | `NUF-210`, `NUF-212` |
| GraphRAG Weaviate store | `internal/data/weaviate/graph` | `NUF-210`, `NUF-212` |
| Vector Weaviate store | `internal/data/weaviate/vector` | `NUF-210`, `NUF-212` |
| PostgreSQL and Supabase adapters | `internal/data/sql/{postgres,supabase}` | `NUF-210`, `NUF-212` |
| Storage contract plus local and GCS implementations | `internal/data/storage`, `internal/data/storage/{local,gcs}` | `NUF-210`, `NUF-212` |
| Canonical/core and legacy task models, executors, planners, and shared contracts/options | `internal/task` | `NUF-212` |
| In-memory/core task services, API support, and adapters | `internal/task/service` | `NUF-212` |
| Canonical/core compatibility bridges and conversion | `internal/task/service/bridge` | `NUF-212` |
| Workflow models and execution | `internal/task/workflow` | `NUF-212` |
| Code/workflow orchestrators, handoffs, and registries | `internal/task/orchestration` | `NUF-212` |
| LLM planning, execution, final synthesis, and routing | `internal/task/orchestration/llm` | `NUF-212` |
| Shared telemetry contracts and fan-out | `github.com/dm-vev/nu/telemetry` | `NUF-201`, `NUF-212` |
| OpenTelemetry/logging and Langfuse | `github.com/dm-vev/nu/telemetry/{otel,langfuse}` | `NUF-201`, `NUF-212` |
| Transport package marker and neutral ownership | `internal/transport` | `NUF-212` |
| Remote wiring, gRPC, HTTP, A2A, and UI transport domains | `internal/transport/{remote,grpc/{client,server,microservice,pb},http/server,a2a/{card,client,server,tool},ui/{server,trace}}` | `NUF-212` |
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

Concrete remote clients belong to `internal/transport/{remote,grpc/client,a2a/client}`, while
`github.com/dm-vev/nu/agent` consumes only transport-neutral contracts supplied
at composition. Generated protobuf belongs to `internal/transport/grpc/pb`; no
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
