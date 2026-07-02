# `internal/app/runtime.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Phase 0 runtime carries injected process IO; real stores/registries are pending.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Shared runtime object passed to mode handlers.

## Code Style

Plain struct with concrete dependencies. Do not hide construction behind
factories unless tests require alternate implementations.

## Types

### `type Runtime struct`

Logic:

- Store fully constructed settings, auth store, resource set, model registry, session store, tool registry, provider registry, and process IO.
- Keep ownership explicit: runtime closes only components it created.
- Do not start goroutines, TUI loops, RPC loops, or provider streams during construction.

Acceptance:

- contains settings, auth store, resource set, model registry, session store,
  tool registry, provider registry, and process IO;
- has no goroutines after construction.

### `type Options struct`

Logic:

- Carry argv, environment, cwd, home, stdin, stdout, stderr, version, clock, and filesystem handles from `cmd/nu`.
- Keep options serializable enough for integration tests to construct without process globals.
- Do not store parsed CLI state here; parsing output belongs to `cli.Request`.

Acceptance:

- carries argv, env, cwd, home, stdin, stdout, stderr, and version metadata.

## Functions

No runtime functions belong in this file. Construction logic stays in
`app.go`; mode dispatch stays in `modes.go`.

Tests:

- `TestRuntimeConstructionHasNoSideEffects`
