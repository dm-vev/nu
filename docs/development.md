# Development

## Workflow

Nu uses Spec Driven Development and TDD:

1. update `spec/functions.md`;
2. re-check `spec/backend.md` for SDK changes or the owning Nu protocol contract;
3. resolve affected `spec/open-questions.md` rows or retain their safe default;
4. add/re-check only the source-file specs needed for this slice;
5. write a failing test;
6. implement and run the narrow test;
7. run `go test ./...` before handing off.

## Commands

```bash
go test ./...
go test -race ./internal/agentui ./internal/session ./internal/rpc ./internal/tui/... ./internal/tools/...
go vet ./...
gofmt -w <files>
```

Use narrow package tests while developing, then the full command.

See [`examples/README.md`](../examples/README.md) for runnable Agent SDK examples.

Useful narrow commands for the UI/RPC slice:

```bash
go test ./internal/agentui ./internal/rpc ./internal/tui/... ./internal/app/...
go test ./internal/agent/... ./internal/contracts ./internal/llm/...
```

Verify the hierarchy migration and examples with:

```bash
go list -f '{{if not .ForTest}}{{.ImportPath}}{{end}}' ./cmd/... ./internal/...
go test ./internal/app/... -run 'TestNUF212Hierarchy|TestNUA009InternalPackages'
go test ./examples/...
```

## Dependency Policy

Nu keeps the dependency graph needed by the complete imported SDK feature set
and Nu. For Nu-owned code, use the standard library first. Add another
dependency only when:

- terminal/platform behavior is not available in stdlib;
- protocol correctness would require too much brittle code;
- the dependency replaces a large amount of code with a stable, small API.

Every new dependency needs a short note in `spec/architecture.md`.

Do not duplicate an imported SDK capability. SDK upgrades and structural
reorganization follow `spec/backend.md` rather than ad-hoc dependency bumps.

## SDK Structure

- Check every SDK package against the exhaustive ownership list in
  `spec/sdk/README.md`; preserve the full imported feature/API behavior and
  owning tests.
- SDK-owned code cannot import Nu-owned packages. Keep application integration
  in Nu composition/adapters rather than adding a compatibility layer.
- Treat packages as domain/dependency boundaries. Move behavior into the exact
  approved owners and delete superseded paths after tests move; do not preserve
  old paths with wrappers.
- Keep roots to shared types/orchestration, use only the exact approved children,
  and never extract a one-helper package.
- Use normal subpackage filenames such as `client.go`; do not repeat package or
  provider names in filenames. Keep all TUI components in `tui/components`.
- Keep one cohesive responsibility per production file. Split files over 300
  lines or record why cohesion requires the exception. Generated/test files are
  exempt.
- Keep a one-file package only when its domain/dependency boundary is real.
- Regenerate protobuf from its `.proto` sources with the command and versions in
  `docs/sdk.md`. Never edit `.pb.go` files or descriptor bytes by hand.

## Test Rules

- No real provider calls in default tests.
- No real `~/.nu` or `~/.pi` reads in tests.
- Use temp directories and explicit env maps.
- Prefer `httptest.Server` for SDK LLM composition tests.
- Prefer fake terminals for TUI.
- Put fixtures in package-local `testdata/`.
- Use `internal/testkit.ScriptedAgent` for TUI/RPC stream tests; do not recreate
  the removed provider event abstraction.

## Documentation Rules

`spec/` describes what must be true. `docs/` describes how to work with the
implementation. If a command, file format, or user-visible behavior changes,
update both when needed.
