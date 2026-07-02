# `internal/agent/agent.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Public session controller for sending prompts, aborting work, and observing
events.

## Code Style

Own concurrency here. Public methods are context-aware and safe for expected TUI
and RPC access.

## Functions

### `New(opts Options) *Agent`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Initialize provider, tools, session store, queues, retry policy, and hooks.
- Start no provider call until `Prompt`.

Acceptance:

- initializes provider, tools, session store, queues, retry policy, and hooks;
- starts no provider call until `Prompt`.

### `(*Agent) Prompt(ctx context.Context, input Prompt) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Reject concurrent prompt without queue behavior.
- Append user message before provider turn.
- Emits agent/turn/message events.

Acceptance:

- rejects concurrent prompt without queue behavior;
- appends user message before provider turn;
- emits agent/turn/message events.

### `(*Agent) Abort()`

Logic:

- Atomically mark the active turn as aborting.
- Cancel the provider stream context and every running tool context owned by the turn.
- Leave queued steering/follow-up messages according to queue policy instead of dropping them blindly.
- Emit abort and queue-state events after cancellation has been requested.

Acceptance:

- cancels active provider stream and running tools;
- preserves queued messages according to mode.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050AbortStopsProviderAndTools`
