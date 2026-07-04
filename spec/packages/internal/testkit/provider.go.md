# `internal/testkit/provider.go`

## Status

Current: IMPLEMENTED
Implementation Commit: a94e00c
Implementation Comments: Scripted provider supports one script per request so agent tool-continuation tests stay deterministic.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Fake provider for agent and integration tests.

## Code Style

Deterministic scripted events. No network. Multiple scripts mean one provider
response per agent request.

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

### `NewScriptedProviderScripts(scripts ...[]provider.Event) *ScriptedProvider`

Logic:

- Copy each request script.
- Serve one script per provider request in order.
- Return a setup error if the agent asks for more scripts than the test supplied.

Acceptance:

- supports multi-request agent loops;
- keeps each recorded request inspectable.

### `NewScriptedProviderErrors(errors []error, scripts ...[]provider.Event) *ScriptedProvider`

Logic:

- Return configured setup errors for matching request indexes before serving scripts.
- Serve scripts only for request indexes that did not return setup errors.

Acceptance:

- Agent retry tests can simulate provider setup errors without network.

### `(*ScriptedProvider) errorCountBefore(index int) int`

Logic:

- Count setup-error slots before a request index so successful requests map to the correct script.

Acceptance:

- Scripts remain ordered even when earlier attempts fail before streaming.

Tests:

- used by `internal/agent` and `internal/app` tests.
