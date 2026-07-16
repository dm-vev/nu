# Testing Spec

## TDD Rule

Nu-owned changes update `NUF-*`/file specs before tests and implementation.
SDK-fork changes additionally update `spec/backend.md` and the ownership-family
index, and keep or adapt owning upstream-style tests for all imported behavior.

Default tests use no real provider, credential, home directory, remote MCP,
database, vector store, or cloud service.

## Test Layers

### NUT-001 Full SDK Baseline Tests

Keep the applicable upstream SDK tests for every feature currently imported from
the pinned baseline. Moves into approved owners may move or adapt tests, but may
not drop feature/API behavior coverage. Tests requiring external infrastructure
remain opt-in or use fakes/emulators.

### NUT-002 Nu Integration Tests

Nu-owned tests prove:

- app auth/model selection creates the right SDK client;
- coding tools in `internal/tools/coding` implement `contracts.Tool`;
- `agentui` maps SDK streams and cancellation;
- print/JSON/RPC/TUI modes retain behavior;
- stdout remains protocol-only.

Use `internal/testkit.ScriptedAgent`, not a recreated provider event protocol.

### NUT-003 Protocol Goldens

Golden tests remain required for session JSONL, RPC JSONL, extension JSONL, TUI
rendering/input, and stable CLI help. SDK provider wire fixtures remain in
their owning `internal/llm/*` tests.

### NUT-004 TUI Tests

Use fixed fake terminals, scripted input, captured frames, resize events, and raw
mode assertions. Rendered lines cannot exceed terminal width.

### NUT-005 Tool Tests

Nu coding tools use temp directories and fake process runners where practical.
Imported SDK tool/MCP adapters keep applicable owning tests. No tool test may
touch user files.

### NUT-006 Race Tests

At minimum race-test Nu concurrency boundaries:

```bash
go test -race ./internal/agentui ./internal/session ./internal/rpc ./internal/tui/... ./internal/tools/...
```

Run imported agent/memory/MCP race tests before upgrading the SDK baseline.

### NUT-007 Import/Attribution Checks

Verification must prove:

- no Go import references `github.com/dm-vev/nu/internal/provider` or `agent-go-sdk`;
- `agent` is the SDK agent package;
- `internal/AGENT_SDK_LICENSE` and `THIRD_PARTY_NOTICES.md` name the pinned source;
- every SDK-owned package belongs to an ownership family in `spec/sdk/README.md`;
- SDK-owned code imports no Nu-owned package;
- production package directories are exactly `cmd/nu` and the root/subpackage/
  standalone allowlist in `spec/sdk/README.md`;
- every domain root owns shared types/orchestration only, with concrete families
  in their approved subpackages;
- all reusable TUI components are in `internal/tui/components`, with no nested
  component package;
- subpackage filenames do not repeat their package/provider prefix;
- import aliases do not encode ancestor domains when the declared package name
  is available; aliases are reserved for real file-local name collisions;
- no production Go package remains at the `internal/` root;
- no legacy compatibility package, alias facade, or forwarding wrapper preserves a
  deleted Nu or upstream SDK path;
- concrete remote clients are owned by `internal/transport/{remote,grpc/client,a2a/client}`, and
  `agent` does not import concrete transport packages;
- `agent` owns the real Agent implementation and cross-domain orchestration,
  imports neither transport nor task, and is not a forwarding facade;
  `agent/{context,config,execution,generation,guardrails,image,mcp,plans,prompts,providers,remote,tools}`
  remain independent domain owners;
- generated protobuf is owned only by `internal/transport/grpc/pb`;
- imported generated protobuf descriptors initialize without panic and a second
  generation produces no diff.

### NUT-008 Security And Output

JSON/RPC stdout contains only protocol records. Active SDK agent warnings route
through the injected logger. Errors and test fixtures do not expose credentials.

### NUT-009 Package And File Structure

Review and CI verification must reject a non-generated production `.go` file
over 300 lines unless its owning `spec/packages/*` file or `spec/backend.md`
records a cohesion-based exception. Generated and test files are excluded.
Package review must also reject one-helper packages and packages that do not
represent a cohesive domain or dependency boundary. Behavior-neutral package
splits retain the same owning tests, not duplicate compatibility tests.

The structural ownership test inventories every production Go package and
compares it to the approved target. It also rejects imports of all superseded
paths, unapproved nested packages, wrappers that merely forward old APIs, and
loss of an imported feature's owning tests. Package count alone is not
feature-retention evidence.

### NUT-010 Agent SDK Examples

`go test ./examples/...` must compile every example. Provider-free examples must
also run without credentials, subprocesses, or network access. The research
example is compile-tested by default and is executed manually only with explicit
provider credentials. Example source must not contain credentials.

## Commands

Develop with the smallest owner set:

```bash
go test ./internal/agentui ./internal/tools/... ./internal/app/... ./internal/rpc ./internal/tui/...
go test ./agent/... ./contracts ./internal/llm/...
go test ./internal/...
```

Hierarchy migration verification:

```bash
go list -f '{{if not .ForTest}}{{.ImportPath}}{{end}}' ./cmd/... ./agent/... ./contracts ./telemetry/... ./internal/...
go test ./internal/app/... -run 'TestNUF212Hierarchy|TestNUA009InternalPackages'
go test ./examples/...
protoc --go_out=. --go_opt=module=github.com/dm-vev/nu \
  --go-grpc_out=. --go-grpc_opt=module=github.com/dm-vev/nu \
  internal/transport/grpc/pb/agent.proto
mv internal/transport/grpc/pb/agent_grpc.pb.go internal/transport/grpc/pb/agentgrpc.pb.go
```

Release verification:

```bash
go test ./...
go test -race ./internal/agentui ./internal/session ./internal/rpc ./internal/tui/... ./internal/tools/...
go vet ./...
```

Commands use an explicit CI timeout. A first SDK build may download the curated
dependency graph; subsequent narrow runs should use the Go cache.

## Evidence

Balanced hierarchy migration evidence is **IMPLEMENTED**. The previous flat
layout's passing tests are baseline evidence only; they do not prove NUA-011.
Completion requires the exact package inventory, full tests, required race/vet
commands, provider-free examples, and no-diff protobuf regeneration from
`internal/transport/grpc/pb/agent.proto`.

An SDK structural change is complete only when the full imported feature/API
behavior and owning test coverage still pass. A Nu integration is implemented
only when a Nu-owned test exercises it through the actual SDK interface.
Documentation, package presence, and compilation alone are not enough for a
user-facing claim.
