# Public Agent SDK

Status: **IMPLEMENTED**. The agent SDK is importable by external Go modules.

The backend is a curated fork sourced from `Ingenimax/agent-sdk-go` `v0.2.62`.
The public SDK boundary is `github.com/dm-vev/nu/agent` with shared contracts
at `github.com/dm-vev/nu/contracts`. Nu-specific providers, storage, memory,
transport, and UI implementations remain private under `internal/`.

Approved owners:

- `github.com/dm-vev/nu/agent`: Agent type, construction, options, and cross-domain orchestration;
- `github.com/dm-vev/nu/agent/context`: context and sub-agent invocation values;
- `github.com/dm-vev/nu/agent/{config,execution,generation,image,mcp,plans,providers,remote,tools}`: cohesive runtime domains;
- `github.com/dm-vev/nu/agent/{config,plans,guardrails,prompts}`: deployment configuration,
  execution plans, guardrails, and prompts;
- `github.com/dm-vev/nu/contracts`: LLM, Tool, Memory, streaming, and transport-neutral
  contracts;
- `internal/config`: SDK configuration not owned by an agent;
- `internal/llm`: shared LLM types, retry, and structured output;
- `internal/llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}`:
  provider clients and streams;
- `internal/memory/{conversation,history,redis,vector,factory}`: conversation
  context/history, Redis, vector retrieval, and config construction;
- `internal/multitenancy`: tenancy;
- `internal/mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}`: MCP domains;
- `internal/tools/{agent,calculator,registry}`: agent-as-tool, Calculator, and
  registry domains;
- `internal/tools/coding`: cwd-scoped filesystem/process tools and `Builtins`;
- `internal/tools/search`, `internal/tools/image/{edit,generation}`, and
  `internal/tools/graphrag`: search integrations, image tools and sessions,
  and GraphRAG adapters;
- `internal/data/embedding` and `internal/data/embedding/{gemini,openai}`:
  shared embedding behavior and provider adapters;
- `internal/data/weaviate/{graph,vector}`, `internal/data/sql/{postgres,supabase}`,
  and `internal/data/storage/{gcs,local}`: retrieval and concrete persistence
  adapters; root `internal/data` is index-only;
- `internal/task`: canonical/core and legacy models, executors, planners, and
  shared task contracts/options;
- `internal/task/service` and `internal/task/service/bridge`: in-memory/core
  services, API support, adapters, and compatibility conversion behavior;
- `internal/task/workflow`: workflow models and execution;
- `internal/task/orchestration` and `internal/task/orchestration/llm`:
  shared/code orchestration plus LLM planning, execution, and routing;
- `github.com/dm-vev/nu/telemetry` plus `{otel,langfuse}`: shared telemetry and integrations;
- `internal/transport/remote`: remote-agent construction;
- `internal/transport/{remote,grpc/{client,server,microservice},http/server,a2a/{card,client,server,tool},ui/{server,trace}}`:
  concrete transport domains;
- `internal/transport/grpc/pb`: generated protobuf only.

`spec/sdk/README.md` is an ownership index. Preserve every feature/API behavior
and owning test currently imported from the pinned baseline. Do not add wrappers
for old paths or duplicate the backend. SDK-owned code must not import Nu-owned
packages.

The migration deletes no feature or API behavior. Callers use final owners
directly; legacy aliases, facade packages, and forwarding wrappers for old paths
are forbidden. Root
packages own shared types/orchestration only. Subpackages use normal filenames
such as `client.go`, and no subpackage exists for one helper. Every TUI component
shares `internal/tui/components`.

Concrete remote construction belongs to the owning transport package; the agent
package depends only on `contracts.RemoteAgentClient`.

Data constructors are package-local and concise: `embedding.NewGemini`,
`graph.NewStore`, `vector.NewStore`, `sql.NewPostgres`,
`sql.NewSupabase`, `storage.NewLocal`, and `storage.NewGCS`.

The target schema and its only generated Go set live in
`internal/transport/grpc/pb`. Regenerate them with `protoc` 3.21.12,
`protoc-gen-go` v1.36.11, and `protoc-gen-go-grpc` v1.6.1 from the repository
root with:

```bash
protoc --go_out=. --go_opt=module=github.com/dm-vev/nu \
  --go-grpc_out=. --go-grpc_opt=module=github.com/dm-vev/nu \
  internal/transport/grpc/pb/agent.proto
mv internal/transport/grpc/pb/agent_grpc.pb.go internal/transport/grpc/pb/agentgrpc.pb.go
```

Treat a package as a domain/dependency boundary and a file as one cohesive
responsibility. Move behavior into the approved owners, delete superseded paths,
and split cohesive files as needed while features and tests remain.
A production file over 300 lines needs a split or documented exception;
generated/test files are exempt. Keep a one-file package only for a real boundary.

When updating the fork, preserve `internal/AGENT_SDK_LICENSE`, update
`THIRD_PARTY_NOTICES.md`, follow `spec/backend.md`, and regenerate imported
protobuf from `.proto` source instead of editing generated Go or descriptor
bytes.
