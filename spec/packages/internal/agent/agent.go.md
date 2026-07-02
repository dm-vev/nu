# `internal/agent/agent.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Prompt, abort, event callback, and Phase 1 fake-tool execution are in scope.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Public controller for sending prompts, aborting work, executing Phase 1 test
tools, and observing events.

## Code Style

Own concurrency here. Public methods are context-aware and safe for expected TUI
and RPC access.

## Functions

### `New(opts Options) *Agent`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Initialize provider settings, event callback, and in-process test tools.
- Start no provider call until `Prompt`.

Acceptance:

- initializes provider settings, event callback, and tools map;
- starts no provider call until `Prompt`.

### `(*Agent) Prompt(ctx context.Context, input Prompt) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Reject concurrent prompt without queue behavior.
- Start one provider turn from the supplied prompt text.
- Emit agent/turn/message events through the callback.

Acceptance:

- rejects concurrent prompt without queue behavior;
- sends one prompt to provider;
- emits agent/turn/message events.

### `(*Agent) Abort()`

Logic:

- Atomically mark the active turn as aborting.
- Cancel the active provider stream context.
- Leave idle agents unchanged.

Acceptance:

- cancels active provider stream.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050AbortStopsProviderAndTools`
