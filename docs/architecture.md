# Nu Architecture

Nu combines its Go CLI/TUI with a curated internal fork of
`Ingenimax/agent-sdk-go` `v0.2.62`.

```text
cmd/nu -> internal/app
            -> app/{auth,cli}
            -> agent/{config,plans,guardrails,prompts}
            -> llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
            -> tools/{coding,search,image,graphrag}
            -> data/{embedding,weaviate/{graph,vector},sql,storage}
            -> task/{service,workflow,orchestration}
            -> telemetry/{otel,langfuse}
            -> transport/{grpc/pb,http,a2a,ui}
            -> internal/agentui     SDK stream to Nu event adapter
                 -> tui/{core,editor,engine,input,message,terminal,components}
                 -> internal/rpc
```

The full feature set currently imported from pinned SDK `v0.2.62` lives directly
under `internal/`: agent runtime and contracts, all imported providers, memory
and retrieval, MCP and tools, orchestration/workflows/tasks, A2A/gRPC/service
integration, guardrails, tracing, and supporting families. No feature is
deleted. There is no nested SDK directory, frozen upstream folder mirror, old
`internal/provider` backend, or wrapper for a superseded import path.

The exact approved roots/subpackages are shown above. Standalone packages are
`agentui`, `config`, `contracts`, `memory`, `multitenancy`, `mcp`, `model`,
`rpc`, `session`, and `testkit`; `cmd/nu` is the only command package. Root
packages own shared types and cross-subpackage orchestration only.

`internal/tools` owns Registry, Calculator, shared helpers, and agent-as-tool
orchestration. Its `coding`, `search`, `image`, and `graphrag` children own their
implementations directly; root does not re-export them. `internal/agent` owns
model/tool execution. `internal/agentui` owns only busy, abort,
reset/model-swap lifecycle and translates SDK events for the existing TUI/RPC.
Nu coding tools and `Builtins(cwd)` stay together in `internal/tools/coding`. TUI runtime
layers use the approved child packages, and every reusable component shares the
single `internal/tui/components` package.

CLI parsing and credentials are Nu-owned in `internal/app/{cli,auth}`;
composition remains in `internal/app`, and model metadata remains in
`internal/model`. Switching models rebuilds the SDK agent
with the same memory and tools. Print/JSON/RPC inject telemetry that cannot write
SDK diagnostics to stdout.

SDK providers live in their `internal/llm/*` packages while shared retry and
structured-output orchestration remain at the root. Data implementations live
only in `internal/data/{embedding,weaviate/{graph,vector},sql,storage}`: embedding
owns generic metadata filtering, Weaviate exposes distinct GraphRAG and vector
stores with focused child packages, SQL owns PostgreSQL/Supabase, and storage
owns its contract plus local/GCS backends.
The root data package is an index, not a facade. Task, telemetry, and transport
implementations follow their listed cohesive packages. Generated
protobuf lives only in `internal/transport/grpc/pb`. Concrete remote clients
live outside `internal/agent`; app injects them through transport-neutral contracts so agent and
transport do not form an import cycle.

Task services and compatibility adapters live in `task/service`, workflow models
and execution live in `task/workflow`, and orchestrators, handoffs, and routers
live in `task/orchestration`. The root keeps canonical/core and legacy models,
executors, planners, and shared task contracts/options; it does not import or
re-export its children.
`internal/agent` does not import `task`.

Within agent, `config` owns independent YAML/deployment/remote configuration,
including loading, merge, environment expansion, persistence, validation, and
conversion. Factories that construct `*agent.Agent` stay in root. `plans` owns
plan models, storage, generation, and execution, while root owns the Agent-facing
plan methods. `guardrails` owns concrete implementations and `prompts` owns
templates, stores, and management. Root imports `config` and `plans`, consumes
guardrails only through `contracts.Guardrails`, and imports neither transport nor
task. Remote clients are injected through `contracts.RemoteAgentClient`.

SDK packages point inward only to other curated SDK owners. They never import
Nu application, UI, or session packages. Nu composition and
`internal/agentui` provide the one-way integration in the other direction.

The balanced migration is **IMPLEMENTED** and intentionally source-breaking.
Superseded paths are deleted after behavior and tests move; old paths get no
aliases, facades, or wrappers. Files inside a subpackage use ordinary names such
as `client.go`, not provider-prefixed names. A subpackage must be cohesive, not a
home for one helper.

See `spec/backend.md` for provenance and local patches, and
`spec/architecture.md` for normative ownership.
