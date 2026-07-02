# `internal/testkit/provider.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Scripted provider exists for agent tests; richer stream scripts can be added when tool calls land.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Fake provider for agent and integration tests.

## Code Style

Deterministic scripted events. No network.

## Functions

### `NewScriptedProvider(events ...provider.Event) *ScriptedProvider`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Emits scripted events in order.
- Record requests.
- Respects context cancellation.

Acceptance:

- emits scripted events in order;
- records requests;
- respects context cancellation.

Tests:

- used by `internal/agent` and `internal/app` tests.
