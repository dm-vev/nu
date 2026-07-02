# Testing Spec

## TDD Rule

For every non-trivial feature:

1. Add or update `NUF-*` requirement.
2. Write a failing test named after the requirement.
3. Implement only enough code to pass.
4. Refactor while tests stay green.

Docs-only changes do not need Go tests.

## Test Layers

### NUT-001 Unit Tests

Use stdlib `testing`. Unit tests run without network, real home directory, real
provider credentials, or long-lived subprocesses.

Required fakes live in `internal/testkit`.

### NUT-002 Golden Tests

Use golden files for stable wire formats:

- provider request JSON;
- provider stream event conversion;
- session JSONL records;
- RPC command/response frames;
- CLI help output when stable.

Golden updates must be intentional and reviewed with the spec diff.

Protocol golden tests are required before code can rely on these contracts:

- `spec/protocols/provider-stream.md`
- `spec/protocols/session-jsonl.md`
- `spec/protocols/rpc-jsonl.md`
- `spec/protocols/extension-jsonl.md`
- `spec/protocols/tui-rendering.md`

### NUT-003 Integration Tests

Integration tests run the real `nu` binary or `internal/app` with temp home and
temp cwd. They use fake providers and fake extension processes by default.

### NUT-004 TUI Tests

TUI tests use a fake terminal:

- fixed width/height;
- scripted key input;
- captured frames;
- resize events;
- raw mode lifecycle assertions.

Tests must assert that rendered lines do not exceed terminal width.

### NUT-005 Provider Contract Tests

Provider tests do not hit real APIs by default. They use `httptest.Server` and
record request bodies, headers, cancellation, retry, and stream parsing.

Real provider smoke tests are opt-in with an env var such as `NU_REAL_API=1`.

### NUT-006 Tool Tests

Built-in tools use temp directories and explicit fake process runners where
possible. Bash tests that spawn real commands must use portable shell snippets or
be platform-scoped.

### NUT-007 Race Tests

Concurrency-heavy packages must pass targeted `go test -race`:

- `internal/agent`;
- `internal/session`;
- `internal/tui`;
- `internal/extension`;
- `internal/tool`.

## Standard Commands

```bash
go test ./...
go test -race ./internal/agent ./internal/session ./internal/extension ./internal/tool
go vet ./...
```

## Spec Drift Checks

Before marking any file spec `IMPLEMENTED`:

- the matching Go file exists;
- every listed test either exists or is replaced by an equivalent named test in
  the file spec;
- protocol golden tests pass when the file touches a protocol;
- `Implementation Commit` is set after commit.

## Fixtures

Use `testdata/` under the package that owns the behavior. Shared fixtures belong
under `internal/testkit/testdata/`.

## Network And Secrets

Default tests must not read real `~/.nu`, `~/.pi`, or process provider keys.
Tests pass explicit env maps and temp paths.
