# `internal/app/app.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Phase 0 dispatch skeleton exists; config/auth/resource wiring is still pending.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Composition root for one Nu process invocation.

## Code Style

Construct dependencies explicitly. Keep side effects behind injected process IO,
filesystem, env, and clock values so tests can run in temp homes.

## Functions

### `Run(ctx context.Context, opts Options) int`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Initialize local state, then enter the smallest required loop.
- Stop on context cancellation, terminal command, or unrecoverable error and clean up owned resources.
- Parses CLI once.
- Resolve paths/settings/auth/resources before mode dispatch.
- Never writes machine-mode diagnostics to stdout.
- Return stable exit codes.

Acceptance:

- parses CLI once;
- resolves paths/settings/auth/resources before mode dispatch;
- never writes machine-mode diagnostics to stdout;
- returns stable exit codes.

### `NewRuntime(ctx context.Context, opts Options) (*Runtime, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Wires config, auth, resources, model registry, tools, sessions, and providers.
- Do not start TUI, RPC, or provider streams.

Acceptance:

- wires config, auth, resources, model registry, tools, sessions, and providers;
- does not start TUI, RPC, or provider streams.

Tests:

- `TestAppRunHelp`
- `TestAppRunPrintModeUsesInjectedRuntime`
