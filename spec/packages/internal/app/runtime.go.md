# `internal/app/runtime.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Runtime carries process IO and an optional agent built from an injected provider.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Shared runtime object passed to mode handlers.

## Code Style

Plain structs with concrete dependencies. Construct the agent directly when a
provider is supplied; do not add registries before there are multiple adapters.

## Types

### `type Runtime struct`

Logic:

- Store normalized process IO and the optional agent used by print mode.
- Keep ownership explicit: runtime closes only components it created.
- Do not start goroutines, TUI loops, RPC loops, or provider streams during construction.

Acceptance:

- contains normalized process IO and optional agent;
- has no goroutines after construction.

### `type Options struct`

Logic:

- Carry argv, environment, cwd, home, stdin, stdout, stderr, version, and optional provider settings from `cmd/nu`.
- Keep options serializable enough for integration tests to construct without process globals.
- Do not store parsed CLI state here; parsing output belongs to `cli.Request`.

Acceptance:

- carries argv, env, cwd, home, stdin, stdout, stderr, version metadata, and optional provider.

## Functions

### `newAgent(opts Options) *agent.Agent`

Logic:

- Return nil when no provider is configured.
- Create `agent.Agent` with provider id, API, model, and provider stream.
- Convert the agent `turn_end` event into one stdout line for print mode.
- Ignore non-final events at this layer.

Acceptance:

- no provider means no agent;
- final turn text is printed once.

Tests:

- `TestAppRunPrintModeUsesInjectedRuntime`
