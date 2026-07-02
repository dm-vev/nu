# Development

## Workflow

Nu uses Spec Driven Development and TDD:

1. update `spec/functions.md`;
2. re-check affected `spec/packages/*` files and make function logic concrete;
3. write a failing test;
4. implement;
5. run the narrow test;
6. run `go test ./...` before handing off.

## Commands

```bash
go test ./...
go test -race ./internal/agent ./internal/session ./internal/tool
go vet ./...
gofmt -w <files>
```

Use narrow package tests while developing, then the full command.

Useful narrow commands for the UI/RPC slice:

```bash
go test ./internal/rpc ./internal/tui ./internal/app ./internal/agent
go test -race ./internal/agent ./internal/rpc ./internal/tui
```

## Dependency Policy

Use the Go standard library first. Add a dependency only when:

- terminal/platform behavior is not available in stdlib;
- protocol correctness would require too much brittle code;
- the dependency replaces a large amount of code with a stable, small API.

Every new dependency needs a short note in `spec/architecture.md`.

## Test Rules

- No real provider calls in default tests.
- No real `~/.nu` or `~/.pi` reads in tests.
- Use temp directories and explicit env maps.
- Prefer `httptest.Server` for provider adapters.
- Prefer fake terminals for TUI.
- Put fixtures in package-local `testdata/`.

## Documentation Rules

`spec/` describes what must be true. `docs/` describes how to work with the
implementation. If a command, file format, or user-visible behavior changes,
update both when needed.
