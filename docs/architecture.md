# Nu Architecture

Nu combines its Go CLI/TUI with a curated internal fork of
`Ingenimax/agent-sdk-go` `v0.2.62`.

```text
cmd/nu -> internal/app
            -> app/{auth,cli}
            -> agent/{context,config,execution,generation,guardrails,image,mcp,plans,prompts,providers,remote,tools}
            -> llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
            -> tools/{coding,search,image/{edit,generation},graphrag}
            -> memory/{conversation,history,redis,vector,factory}
            -> mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}
            -> data/{embedding/{gemini,openai},weaviate/{graph,vector},sql/{postgres,supabase},storage/{gcs,local}}
            -> task/{service/{bridge},workflow,orchestration/{llm}}
            -> telemetry/{otel,langfuse}
            -> transport/{remote,a2a/{card,client,server,tool},grpc/{client,server,microservice,pb},http/server,ui/{server,trace}}
            -> internal/agentui     SDK stream to Nu event adapter
                 -> tui/{core,editor,engine,input,message,terminal,components}
                 -> internal/rpc
```

The public SDK surface currently imported from pinned SDK `v0.2.62` lives in
`github.com/dm-vev/nu/agent`, `github.com/dm-vev/nu/contracts`, and
`github.com/dm-vev/nu/telemetry`. Provider, memory, retrieval, MCP, tools,
transport, and UI implementations remain under `internal/`. No feature is
deleted. There is no nested SDK directory, frozen upstream folder mirror, old
`internal/provider` backend, or wrapper for a superseded import path.

The exact approved roots/subpackages are shown above. Transport-neutral remote
construction lives in `internal/transport/remote`; the `transport` root is only
the package marker. Public packages are `agent` and its cohesive child domains, `contracts`, and
`telemetry`; private standalone packages are `internal/agentui`,
`internal/config`, `internal/multitenancy`, `internal/model`, `internal/rpc`,
`internal/session`, and `internal/testkit`. `cmd/nu` is the only command package.
The root `agent` package owns the real Agent implementation and cross-domain
orchestration. Agent-specific adapters that require private Agent state stay in
that package; reusable domains live in cohesive child packages.

`internal/tools/{agent,calculator,registry}` own agent-as-tool, Calculator, and
registry domains. Its `coding`, `search`, `image`, and `graphrag` children own
their implementations directly; there is no root tools facade. `agent`
owns the Agent type and cross-domain orchestration; its configuration, generation,
planning, MCP, provider, remote, and tool domains live in dedicated child
packages. Agent-specific memory, GraphRAG, sub-agent, task, and validation
adapters stay beside the Agent because they use its private state.
`internal/agentui` owns only busy, abort,
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
only in `internal/data/{embedding/{gemini,openai},weaviate/{graph,vector},sql/{postgres,supabase},storage/{gcs,local}}`:
embedding owns shared contracts and generic metadata filtering while provider
clients live in child packages; Weaviate exposes distinct GraphRAG and vector
stores; SQL and storage backends also own separate child packages.
The root data package is an index, not a facade. Task, telemetry, and transport
implementations follow their listed cohesive packages. Generated
protobuf lives only in `internal/transport/grpc/pb`. Remote-agent wiring lives
in `internal/transport/remote`; protocol clients and servers live in their
transport domain packages. Concrete remote clients stay outside
`agent`; app injects them through transport-neutral contracts so
agent and transport do not form an import cycle.

Task services and adapters live in `task/service`, compatibility bridges live
in `task/service/bridge`, workflow models
and execution live in `task/workflow`, shared/code orchestration lives in
`task/orchestration`, and LLM orchestration lives in `task/orchestration/llm`.
The root keeps canonical/core and legacy models,
executors, planners, and shared task contracts/options; it does not import or
re-export its children.
`agent` does not import `task`.

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
