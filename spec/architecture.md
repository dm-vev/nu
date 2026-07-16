# Architecture Spec

## Topology

```text
cmd/nu
  -> internal/app                    process composition/orchestration
     -> internal/app/auth            Nu credential storage/resolution
     -> internal/app/cli             CLI parsing/help/request types
     -> internal/model               Nu model registry
     -> internal/agent               SDK agent runtime orchestration
        -> config plans guardrails prompts
     -> internal/llm                 shared LLM types, retry, structured output
        -> openai anthropic gemini azureopenai deepseek ollama vllm
     -> internal/tools               tool domains
        -> agent calculator registry coding search image graphrag
     -> internal/memory              memory domains
        -> conversation history redis vector factory
     -> internal/mcp                 MCP domains
        -> builder client config fault lazy preset prompt registry resource retry sampling schema tool transport
     -> internal/data                data package index only
        -> embedding weaviate sql storage
     -> internal/task                task models, executor, planner, contracts/options
        -> service workflow orchestration
     -> internal/telemetry           shared telemetry types/orchestration
        -> otel langfuse
     -> internal/transport           shared transport types/orchestration
        -> remote
        -> grpc/{client,server,microservice} -> pb
        -> http/server a2a/{card,client,server,tool} ui/{server,trace}
     -> internal/agentui             SDK stream -> Nu TUI/RPC lifecycle
        -> internal/tui              TUI application orchestration
           -> core editor engine input message terminal components
        -> internal/rpc
```

The imported upstream SDK source lives directly under `internal/`; there is no
`internal/agent-go-sdk` nesting. Upstream package relationships are not an API:
behavior moves into the exact approved hierarchy, superseded paths are deleted,
and files may split while all imported feature/API behavior and tests are
preserved. Imports use the resulting `nu/internal/*` owners; compatibility
wrappers for old upstream paths are forbidden.

The approved breaking target is exhaustive:

```text
app/{auth,cli}
agent/{config,plans,guardrails,prompts}
llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}
tools/{agent,calculator,registry,coding,search,image,graphrag}
memory/{conversation,history,redis,vector,factory}
mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}
data/{embedding,weaviate/{graph,vector},sql,storage}
task/{service,workflow,orchestration}
telemetry/{otel,langfuse}
transport/{remote,grpc/{client,server,microservice,pb},http/server,a2a/{card,client,server,tool},ui/{server,trace}}
tui/{core,editor,engine,input,message,terminal,components}

standalone: agentui config contracts multitenancy model rpc session testkit
```

Every path is below `internal/`; `cmd/nu` remains the only command package. A
listed owner is a package boundary. Family roots such as `transport/grpc` are
index-only directories; protocol behavior belongs in the listed domain package.
The hierarchy is not a compatibility layer: old paths are deleted as callers
move, with no alias facade, forwarding wrapper, or duplicate tree. No feature
or API behavior may be deleted.

### Approved Ownership Map

| Owner | Responsibility |
|---|---|
| `internal/app` | process composition and mode orchestration only |
| `internal/app/auth`, `internal/app/cli` | credentials; CLI parsing/help/request handling |
| `internal/agent` | shared Agent runtime types and run orchestration |
| `internal/agent/{config,plans,guardrails,prompts}` | cohesive agent policy/configuration families |
| `internal/llm` | shared LLM contracts, retry, and structured-output orchestration |
| `internal/llm/{openai,anthropic,gemini,azureopenai,deepseek,ollama,vllm}` | provider clients and provider-specific streaming; OpenAI-compatible variants stay with `openai`, and Claude-on-Bedrock stays with `anthropic` |
| `internal/tools/{agent,calculator,registry}` | agent-as-tool, Calculator, and registry domains |
| `internal/tools/coding` | all seven cwd-scoped Nu filesystem/process tools and `Builtins(cwd)` |
| `internal/tools/search` | WebSearch, GitHub content, and HuggingFace integrations |
| `internal/tools/image` | image generation/edit tools and their sessions |
| `internal/tools/graphrag` | GraphRAG tool adapters |
| `internal/data` | concise package index; no implementation or forwarding API |
| `internal/data/embedding` | embedders plus generic metadata filters and in-memory evaluation |
| `internal/data/weaviate/graph` | GraphRAG `Store` implementation and graph helpers |
| `internal/data/weaviate/vector` | Vector `Store` implementation and metadata helpers |
| `internal/data/sql` | PostgreSQL and Supabase adapters with `Postgres*` and `Supabase*` names |
| `internal/data/storage` | `Storage` contract plus `Local*` and `GCS*` implementations |
| `internal/task` | canonical/core and legacy task models, executors, planners, and shared task contracts/options |
| `internal/task/service` | in-memory/core services, API support, service adapters, and compatibility conversion behavior |
| `internal/task/workflow` | workflow models and execution behavior |
| `internal/task/orchestration` | LLM/code/workflow orchestrators, handoffs, registries, and routers |
| `internal/telemetry` | shared telemetry contracts and fan-out orchestration |
| `internal/telemetry/{otel,langfuse}` | OpenTelemetry/logging and Langfuse integrations |
| `internal/transport` | transport package marker and transport-neutral ownership |
| `internal/transport/remote` | remote-agent construction and gRPC client injection |
| `internal/transport/grpc/client` | remote-agent gRPC client and stream callbacks |
| `internal/transport/grpc/server` | gRPC agent service and protocol handlers |
| `internal/transport/grpc/microservice` | local agent microservice lifecycle and management |
| `internal/transport/a2a/{card,client,server,tool}` | A2A card, client, server, and remote-tool domains |
| `internal/transport/http/server` | HTTP agent endpoints and SSE responses |
| `internal/transport/ui/{server,trace}` | UI HTTP server and trace collection |

The agent split is directional: root `agent` owns every factory that constructs
an `*agent.Agent` and imports `agent/config` and `agent/plans`. `agent/config`
owns YAML, deployment, remote-configuration, MCP-configuration, merge,
environment, persistence, validation, and conversion behavior without importing
root agent or transport. `agent/plans` owns plan models, storage, generation, and
execution without importing root agent. `agent/guardrails` owns concrete
guardrails; root consumes only `contracts.Guardrails`. `agent/prompts` owns prompt
templates, stores, and management. Concrete remote clients remain outside agent
and are injected through `contracts.RemoteAgentClient`; neither root nor a child
imports `internal/transport` or `internal/task`.
| `internal/transport/grpc/pb` | generated protobuf and its `.proto` source only |
| `internal/tui` | terminal application state, slash dispatch, and cross-subpackage orchestration |
| `internal/tui/{core,editor,engine,input,message,terminal}` | cohesive TUI runtime layers |
| `internal/tui/components` | every reusable TUI component; no component-per-package tree |
| `internal/memory/{conversation,history,redis,vector,factory}` | conversation context/history, Redis, vector retrieval, and config construction |
| `internal/mcp/{builder,client,config,fault,lazy,preset,prompt,registry,resource,retry,sampling,schema,tool,transport}` | MCP client, transports, retry, lazy servers, management, and protocol domains |
| standalone packages | `agentui`, `config`, `contracts`, `multitenancy`, `model`, `rpc`, `session`, `testkit` |

Files inside a subpackage use ordinary responsibility names such as `client.go`,
`stream.go`, and `client_test.go`; they do not repeat the package/provider name.
A subpackage must contain a cohesive feature boundary, not a single helper
extracted to satisfy a size or naming preference.

## Ownership

| Concern | Owner |
|---|---|
| Model/tool agent loop | `internal/agent` SDK fork |
| LLM clients and provider streaming | `internal/llm/*` SDK fork |
| Retry and structured-output orchestration | `internal/llm` |
| Conversation memory and MCP client support | `internal/memory/{conversation,history,redis,vector,factory}`, `internal/mcp/*` |
| Data and retrieval | `internal/data/{embedding,weaviate/{graph,vector},sql,storage}` |
| SDK diagnostics and tracing | `internal/telemetry`, `internal/telemetry/*` |
| Nu process composition, auth/CLI, and model selection | `internal/app`, `internal/app/*`, `internal/model` |
| SDK tools and Nu coding tools | `internal/tools/{agent,calculator,registry,coding,search,image,graphrag}` |
| TUI/RPC busy, abort, event translation | `internal/agentui` |
| Terminal UI and slash commands | `internal/tui`, `internal/tui/*` |
| Local JSONL RPC | `internal/rpc` |
| Branchable local sessions and compaction | `internal/session` |
| A2A, gRPC, HTTP/service, UI transport, and remote clients | `internal/transport/{remote,grpc/*,http/server,a2a/*,ui/*}` |
| Generated transport protobuf | `internal/transport/grpc/pb` |

One concern has one runtime owner. `internal/agentui` cannot invoke an LLM or a
tool; `internal/app` cannot implement a model loop; Nu coding tools cannot call
providers.

## Imported Backend

Source baseline:

- repository: `https://github.com/Ingenimax/agent-sdk-go`
- tag: `v0.2.62`
- commit: `71a421c747c64392fd71e665067393a65e188bf7`
- license: MIT, preserved in `internal/AGENT_SDK_LICENSE`

The source and feature baseline is the complete upstream `pkg/` tree currently
imported from the pinned version, including all API behavior and owning tests.
`spec/sdk/README.md` records current ownership families rather than selecting
capabilities. Upstream CLI, examples, repository CI, and branding assets are
outside that imported SDK baseline. Generated protobuf must be regenerated from
its `.proto` sources with the documented toolchain after package or schema
changes. Hand-editing `.pb.go` files, descriptor bytes, or generated import paths
is forbidden.

Nu modifications to SDK-owned packages must preserve the complete imported
feature/API behavior and be recorded in `spec/backend.md`. Moves into the
approved owners, deletion of superseded paths, and behavior-neutral file splits
are permitted; compatibility shims are not. SDK-owned packages may import other
SDK-owned packages, the standard library, and declared external dependencies,
but must not import Nu-owned packages. Nu integration points depend on the SDK
in the opposite direction.

## Runtime Flow

### Startup

1. CLI parses a typed request.
2. Nu auth/settings/model registry select provider and model.
3. `internal/app` constructs a provider client from the selected `internal/llm/*` package.
4. Nu coding tools in `internal/tools/coding` implement `contracts.Tool`.
5. `internal/app` constructs SDK `agent.Agent` with memory, tools, stream config,
   bounded tool iterations, and a stdout-safe logger.
6. `internal/agentui` receives the SDK `StreamingAgent` and dispatches to
   print, JSON, RPC, or TUI mode.

### Prompt

1. `agentui.Controller` rejects a concurrent UI prompt and owns cancellation.
2. It supplies stable Nu organization/conversation IDs required by SDK memory.
3. SDK `Agent.RunStream` owns model calls, tool execution, retries, conversation
   memory, and configured MCP calls.
4. `agentui` translates SDK stream events without changing backend behavior.
5. TUI/RPC consume existing Nu events; print mode emits final accumulated text.

### Model Switch

SDK LLM clients bind their model at construction. Nu rebuilds SDK Agent with the
new LLM and the same memory/tool set, then atomically swaps the idle UI
controller. Active prompts cannot switch model.

## Package Families

The approved roots, subpackages, and standalone packages are exactly the target
listed under Topology. Changing that allowlist requires a new architecture
decision. The imported feature/API behavior set may not shrink. Nu need not
expose every imported SDK feature in its CLI.

## Architecture Decisions

### NUA-001: Process Extensions

Nu-specific extensions remain out-of-process. SDK MCP is the preferred standard
tool/resource/prompt protocol; Nu JSONL extensions remain for TUI/lifecycle jobs.

### NUA-002: Native Session Format With Import

Nu branchable JSONL sessions remain application-owned. SDK memory powers active
conversation context until a session-memory adapter is connected.

### NUA-003: Feature-Complete Curated Upstream SDK Fork

The pinned SDK is the source baseline for a curated fork directly under
`internal/`, not a frozen folder-topology mirror. Preserve every feature and API
behavior currently imported from the pinned baseline while allowing structural
reorganization under NUA-009 and the approved target in NUA-011.

### NUA-004: TUI Anti-Corruption Adapter

Keep SDK event/API details out of terminal components through one small
`internal/agentui` adapter. It owns UI concurrency/cancel semantics only.

### NUA-005: No Legacy Backend

`internal/provider` and the previous custom agent loop are deleted. Compatibility
code cannot recreate provider request/event abstractions or preserve removed SDK
package paths in another directory.

### NUA-006: Preserve Upstream Attribution

MIT notice and pinned source identity ship with every substantial source copy.
Rebranding package paths does not remove upstream copyright.

### NUA-007: Protocol Stdout Isolation

Nu injects a non-stdout SDK logger in print/JSON/RPC/TUI composition. Direct
printing in active SDK agent paths is replaced by logger calls.

### NUA-008: Upstream-First Maintenance

Fix defects in imported SDK behavior with a focused test and track the patch.
Upstream remains the source reference. Its exact folder and import-path topology
is not a maintenance obligation, but the imported feature/API behavior is.

### NUA-009: Package And File Boundaries

A package represents one domain or dependency boundary, not an upstream folder,
feature backlog, or file-size workaround. The NUA-011 target is exhaustive;
adding, removing, or merging a target package requires a new architecture
decision. A one-file package is allowed only when it is a cohesive domain or
dependency boundary, never when it contains one extracted helper.

A production file owns one cohesive responsibility. A behavior-neutral split
within the same package is allowed. A non-generated production `.go` file over
300 lines must be split or have a documented exception in its owning package
spec or `spec/backend.md`, including why a split would reduce cohesion.
Generated files and test files are excluded from the 300-line rule.

### NUA-010: Fork Dependency Direction And Generated Source

SDK-owned code cannot import Nu-owned packages. Nu composition and adapters may
import SDK-owned packages. Generated protobuf is changed only by updating its
source/schema or package plan and running the documented generator; generated
Go or descriptor data is never edited by hand.

### NUA-011: Balanced Package Hierarchy

Migrate from the temporary flat-root consolidation to the exhaustive balanced
hierarchy above. This decision is source-breaking by design: callers update
imports directly, and superseded paths are deleted without wrappers. Root
packages keep only shared types/orchestration; cohesive implementation families
belong in the listed subpackages. All TUI components share one `components`
package. Generated protobuf belongs only to `transport/grpc/pb`. No feature or
API behavior is deleted.

## Dependencies

`go.mod` retains dependencies required by Nu and the complete imported SDK
feature set. A dependency may be removed after structural reorganization only
when no preserved feature or owning test needs it. New dependencies still
require an architecture decision.
