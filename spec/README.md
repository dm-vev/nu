# Nu Canonical Specification

Status: **agent SDK backend integrated; application features partially implemented**.

Nu is a local-first coding agent. Pi is the application UX baseline. A curated
fork sourced from `Ingenimax/agent-sdk-go` `v0.2.62` is the internal backend and
lives directly under `internal/`.

## Precedence

1. `product.md` for product boundary/security/non-goals.
2. `architecture.md` and accepted `NUA-*` decisions.
3. `backend.md` for imported SDK provenance, transformation, and patch ledger.
4. Owning `protocols/*`, then `functions.md` and `testing.md`.
5. `examples.md` for runnable Agent SDK examples.
6. `packages/*` for Nu-owned source-file contracts.
7. `docs/*` for contributor guidance.

The full imported SDK feature/API behavior is owned by source plus owning tests.
`sdk/README.md` indexes ownership families, not allowed capabilities; upstream
folder/import-path compatibility is not a requirement. Nu-owned adapters and
modifications require Nu specs.

## Reading Order

1. [Product](product.md), [architecture](architecture.md), and
   [backend provenance](backend.md).
2. [Capability matrix](capabilities.md) and [functional requirements](functions.md).
3. [SDK package index](sdk/README.md) and [agent run flow](flows/agent-run.md).
4. [Protocols](protocols/README.md), [testing](testing.md), and
   [implementation status](implementation-status.md).

## Canonical Owners

| Subject | Owner |
|---|---|
| Product scope | `product.md` |
| Runtime/package boundaries | `architecture.md` |
| SDK source/version/reorganization/patches | `backend.md` |
| User-visible requirements | `functions.md` |
| SDK exact Go API | `internal/*` source and owning tests |
| Runnable SDK examples | `examples.md` |
| TUI/RPC SDK adaptation | `packages/internal/agentui/*` |
| Persisted/wire formats | `protocols/*` |
| Verification | `testing.md` |

## Status Semantics

- `IMPLEMENTED`: source exists and owning tests pass.
- `IN_PROGRESS`: the approved target is being implemented and current source may
  still reflect a documented temporary layout.
- `PARTIAL`: usable source exists but its product integration is incomplete.
- `SPECIFIED`: implementation-ready target without source evidence.
- `PROVISIONAL`: safe default selected pending resolution.
- `BLOCKED`: implementation cannot proceed.
- `REMOVED`: superseded source is absent and must not return.
- `RESOLVED` / `SUPERSEDED`: retained question/decision history.

Every `NUA-*` heading is accepted unless marked otherwise.

## IDs

- `NUF-*`: functional requirements.
- `NUT-*`: testing requirements.
- `NUA-*`: architecture decisions.
- `NUQ-*`: cross-package questions.

IDs are stable and never reused.

## Workflow

1. Update canonical requirement/architecture.
2. For SDK code, update `backend.md` and the ownership-family index; for Nu code,
   update the owning file spec.
3. Write the smallest failing test.
4. Implement without recreating an existing SDK capability.
5. Run narrow tests, imported owner tests, race tests, then full verification.
6. Update implementation evidence and attribution when the source baseline moves.

## Compatibility

Pi compatibility means application behavior/data import, not TypeScript source.
The internal SDK is a modified, curated MIT fork. Upstream folder and import-path
compatibility are not promised; imported feature and API behavior remain
available through the resulting owners.
