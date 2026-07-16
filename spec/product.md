# Product Spec

## Goal

Nu is a local-first coding agent with a Go TUI and a curated agent SDK backend.
The user-facing artifact is the `nu` binary. External consumers use the public
agent API and implementation at `github.com/dm-vev/nu/agent`
and shared contracts at
`github.com/dm-vev/nu/contracts`; model providers, conversation memory,
retry/logging, and MCP implementations remain under `internal/`.

Nu owns the product shell: CLI/TUI, local auth/settings, coding tools, branchable
sessions, resources, Pi compatibility, RPC/JSON modes, and exports. It does not
maintain a second model/tool agent loop.

## Baselines

Pi remains the application UX baseline:

- interactive coding-agent CLI and TUI;
- provider/model/auth management;
- read/write/edit/bash/grep/find/ls tools;
- session resume/fork/clone/tree and compaction;
- slash commands, resources, JSON/RPC, export/share/update.

`Ingenimax/agent-sdk-go` `v0.2.62` is the source and feature baseline. Every SDK
feature and API behavior currently imported from that baseline remains
available, whether or not Nu exposes it in the CLI. `spec/sdk/README.md` records
current ownership families. It does not freeze upstream folder topology:
behavior moves into NUA-011 owners and cohesive files may split when behavior
and tests remain intact.

The user-approved balanced hierarchy is intentionally source-breaking and is
defined exhaustively by NUA-011. Domain roots own shared types/orchestration;
cohesive implementation families live in the listed subpackages. No feature or
API behavior is deleted. Old import paths are removed without aliases or wrappers.

## Distribution

The primary artifact is one `nu` binary. The public SDK module path is
`github.com/dm-vev/nu`; its stable agent entry point is
`github.com/dm-vev/nu/agent`. Private Nu integrations remain behind `internal/`.

No Node/TypeScript runtime is required. MCP/extension subprocesses may use any
runtime selected explicitly by the user.

## Non-Goals

- No parallel home-grown provider abstraction or agent loop.
- No compatibility wrappers or duplicate backend for moved, merged, renamed, or
  internalized SDK package paths.
- No source compatibility with Pi TypeScript.
- No source or package-layout compatibility promise for the upstream SDK.
- No removal of upstream MIT attribution.
- No silent telemetry or cloud control plane required for local operation.
- No automatic Nu CLI enablement of SDK services, remote config, datastores, or
  network listeners without an accepted Nu requirement; their imported SDK
  library behavior remains available.

## Persistence Paths

- global config: `~/.nu/agent/`
- project config: `.nu/`
- sessions: `~/.nu/agent/sessions/`
- auth: `~/.nu/auth.json`
- trust: `~/.nu/agent/trust.json`

SDK library mode receives no implicit persistence from Nu. The application
composition root supplies memory/config explicitly.

## Security Baseline

Nu tools execute with the user's OS permissions. Project resources and
executable extensions require trust. Provider credentials are resolved by Nu
and passed to SDK constructors; they must not enter prompts, events, session
records, or protocol stdout. Headless stdout remains protocol-only even when SDK
logging is enabled elsewhere.

## Done Definition

A backend migration is done only when:

- `github.com/dm-vev/nu/agent` is the integrated SDK package;
- no `internal/provider` or custom agent loop remains;
- TUI/RPC operate through `internal/agentui` and SDK stream events;
- Nu coding tools implement the SDK Tool interface;
- the full imported SDK feature/API behavior and owning test coverage remain
  available after any package reorganization, and SDK-owned code imports no
  Nu-owned package;
- generated protobuf is reproducible from its `.proto` sources and is not
  hand-edited, and lives in `internal/transport/grpc/pb`;
- only the approved target packages remain, with remote-client implementation
  outside `agent` and no compatibility wrappers;
- upstream attribution is present;
- SDK ownership tests and Nu app/TUI/RPC tests pass.
