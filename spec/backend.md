# Integrated Agent SDK Backend

Status: **IMPLEMENTED: full baseline imported and balanced hierarchy applied**.

## Provenance

| Field | Value |
|---|---|
| Upstream | `Ingenimax/agent-sdk-go` |
| Version | `v0.2.62` |
| Commit | `71a421c747c64392fd71e665067393a65e188bf7` |
| Source review subtree | `pkg/` |
| Nu destination | `internal/` |
| License | MIT |

The pinned version identifies both source provenance and the feature baseline.
Every SDK feature and API behavior currently imported from that baseline remains
available. Exact upstream package paths and folder topology are not compatibility
requirements.

## Retention Rule

Retain the complete feature and API-behavior set currently imported from the
pinned upstream `pkg/` baseline, together with tests that cover that full set.
`sdk/README.md` records ownership families without limiting the feature baseline.
Behavior moves into the exact NUA-011 owners, and cohesive files may split. A
superseded package path may be deleted only after its
behavior and tests move to the resulting owner. Do not duplicate the backend or
add compatibility wrappers solely to preserve an old path.

## Mechanical Transformation

1. Start from the complete pinned upstream `pkg/` source currently imported
   under `internal/`, including its feature/API behavior and owning tests.
2. Move or split packages only when all behavior and test coverage transfer to
   the approved balanced owner; delete only the superseded path.
3. Rewrite imports directly to the final `nu/internal/*` owners. Do not preserve
   old paths with wrappers or a second compatibility backend.
4. Permit behavior-neutral in-package file splits and record package-level
   structural changes in this ledger.
5. Remove a module dependency only when the reorganized full feature set and its
   tests no longer reference or require it.
6. Regenerate required protobuf from its `.proto` sources after schema,
   package, or import changes. Never hand-edit generated Go or descriptors.
7. Upstream repository material outside the imported SDK source baseline, such
   as its CLI, examples, docs site, CI, and image assets, need not be copied.

Any imported generated protobuf requires its `.proto` source, generator command,
and pinned `protoc` and Go plugin versions to be recorded here before generated
output changes. If the pinned SDK subtree lacks the source schema, obtain it from
the pinned upstream commit; do not patch the generated output instead.

## Nu Patch Ledger

| Area | Nu change | Reason |
|---|---|---|
| `internal/agent/streaming.go` | warnings use injected logger | preserve JSON/RPC stdout |
| import paths | rewrite callers directly to each approved `nu/internal/*` owner | internal monorepo integration without wrappers |
| temporary flat layout | consolidated behavior into domain roots and removed legacy wrappers | completed intermediate migration step; not the approved final topology |
| balanced hierarchy | split the flat roots into the exact roots/subpackages in NUA-011; roots retain shared types/orchestration only | **IMPLEMENTED**; cohesive boundaries restored without feature loss or wrappers |
| app | move credential behavior to `app/auth` and CLI parsing/help/request behavior to `app/cli` | keep `app` as process composition/orchestration |
| agent | move deployment config, plans, guardrails, and prompts to `agent/{config,plans,guardrails,prompts}` | keep agent runtime shared types/orchestration at the root |
| LLM | move provider implementations to the seven approved `llm/*` packages | use ordinary names such as `client.go`; retain retry/structured-output orchestration at the root |
| tools and data | move cohesive implementations to `tools/{coding,search,image,graphrag}` and `data/{embedding,weaviate/{graph,vector},sql,storage}` | avoid both one-helper packages and an unrelated flat implementation package |
| task | keep models/executors/planners at the root; move services/adapters, workflows, and orchestrators/handoffs/routers to `task/{service,workflow,orchestration}` | completed with direct imports, ordinary filenames, and no root facade |
| telemetry | move implementations to `telemetry/{otel,langfuse}` | preserve all imported diagnostics behavior |
| transport | move concrete families to `transport/{grpc,http,a2a,ui}` while retaining transport-neutral orchestration at the root | retain remote behavior without an `agent` import cycle |
| generated protobuf | move schema and generated output from `internal/transportpb` to `internal/transport/grpc/pb` by regeneration | **IMPLEMENTED**; target owns all transport protobuf and generated Go is never hand-edited |
| TUI | move runtime layers to `tui/{core,editor,engine,input,message,terminal}` and all reusable components to one `tui/components` package | remove prefixed flat filenames without recreating component-per-package fragmentation |

## Generated Transport Provenance

The pre-migration checked-in generated headers record `protoc` 3.21.12,
`protoc-gen-go` v1.36.11, and `protoc-gen-go-grpc` v1.6.1. They identify
`internal/transportpb/agent.proto` as their current source. During this migration,
move the schema to `internal/transport/grpc/pb/agent.proto`, update `go_package`,
and regenerate from the repository root with:

```bash
protoc --go_out=. --go_opt=module=nu \
  --go-grpc_out=. --go-grpc_opt=module=nu \
  internal/transport/grpc/pb/agent.proto
mv internal/transport/grpc/pb/agent_grpc.pb.go internal/transport/grpc/pb/agentgrpc.pb.go
```

Completion requires a second target-path generation to produce no diff and all
protobuf descriptor tests to pass.

## Integration Boundary

`internal/app` constructs SDK LLM/Agent instances and concrete remote transports.
`internal/agentui` translates events and owns only prompt
busy/cancel/reset/model-swap UI behavior. Nu coding tools live in
`internal/tools/coding` and implement `contracts.Tool`. No other backend adapter is
permitted.

SDK-owned code must not import `internal/app`, `internal/agentui`,
`internal/model`, `internal/tui`, `internal/rpc`, `internal/session`,
`internal/testkit`, or other Nu-owned packages. Shared behavior needed by SDK
code belongs in an SDK-owned target package or stays at the Nu composition
boundary; it is not bridged with a compatibility package. `internal/agent` must
not import concrete transport subpackages; app composition injects the remote
client through `internal/contracts`.

## Update Procedure

1. Review upstream release notes and license.
2. Diff the complete upstream `pkg/` feature/API and test baseline against the
   pinned commit; do not limit review to features currently exposed by Nu.
3. Reapply import/package transformations in a clean worktree while preserving
   the full imported feature set.
4. Reapply each Nu patch ledger row and resolve conflicts explicitly.
5. Regenerate imported protobuf output when its source, package, or imports
   change and verify the worktree is clean after a second generation.
6. Run the full imported SDK test set, Nu integration tests, structural checks,
   race tests, and vet.
7. Update this file, `sdk/README.md`, `THIRD_PARTY_NOTICES.md`, dependency files,
   and implementation evidence.
