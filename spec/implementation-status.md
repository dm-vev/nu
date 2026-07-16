# Implementation Status

Overall balanced hierarchy migration: **IMPLEMENTED**. The full imported feature
baseline now lives in the approved NUA-011 roots/subpackages. No feature was
deleted, and no compatibility wrapper was added for a superseded path.

| Subsystem | Status | Evidence |
|---|---|---|
| Full imported SDK baseline | IMPLEMENTED | complete upstream `pkg/` feature source and owning tests are exposed through public SDK packages and private implementations |
| Temporary flat-root consolidation | IMPLEMENTED | current source/tests provide the migration baseline; this layout is superseded as the final target |
| Balanced hierarchy migration | IMPLEMENTED | exact NUA-011 production package inventory and direct imports are in place |
| Provenance/license | IMPLEMENTED | `spec/backend.md`, `THIRD_PARTY_NOTICES.md`, `internal/AGENT_SDK_LICENSE` |
| SDK import rewrite | IMPLEMENTED | callers use the final NUA-011 owners directly |
| SDK dependency-direction check | IMPLEMENTED | structural tests cover approved child packages and keep `agent` independent of concrete transports |
| Generated protobuf relocation/regeneration | IMPLEMENTED | schema and generated output live in `internal/transport/grpc/pb` |
| SDK agent runtime | IMPLEMENTED | `agent`, upstream tests |
| Nu SDK construction | IMPLEMENTED | responsibility files under `internal/app` |
| TUI/RPC stream adapter | IMPLEMENTED | `internal/agentui`, tests |
| Nu app/auth/CLI split | IMPLEMENTED | auth and CLI behavior live in `internal/app/{auth,cli}`; root keeps composition/orchestration |
| TUI balanced split | IMPLEMENTED | runtime layers live in approved child packages and reusable components share `internal/tui/components` |
| Nu and SDK tools split | IMPLEMENTED | root owns Registry/Calculator/agent orchestration; coding, search, image, and GraphRAG live in their cohesive child packages |
| SDK task split | IMPLEMENTED | root models/executors/planners plus direct `service`, `workflow`, and `orchestration` owners; structural tests prohibit root child imports and agent task imports |
| SDK domain hierarchy | IMPLEMENTED | agent, LLM, data, task, telemetry, and transport use the exact cohesive NUA-011 packages |
| Normal subpackage filenames | IMPLEMENTED | production Go filenames contain no underscore separators; Go test suffixes remain conventional |
| Old Nu provider backend | REMOVED | no `internal/provider` source/import |
| Old Nu agent loop | REMOVED | `agent` is upstream package |
| Print/JSON/RPC/TUI | IMPLEMENTED | Nu integration tests |
| Imported SDK feature/test retention | IMPLEMENTED | all package tests pass under final owners |
| Runnable Agent SDK examples | IMPLEMENTED | examples compile against the balanced APIs |
| Session-backed SDK memory | PARTIAL | SDK buffer connected; branchable store remains separate |

Documentation alone is not evidence. SDK rows require source, provenance, and
their owning package tests. Nu integration rows require tests outside the SDK
package.
