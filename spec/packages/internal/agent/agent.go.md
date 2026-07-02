# `internal/agent/agent.go`

## Status

Current: IMPLEMENTED
Implementation Commit: pending
Implementation Comments: Prompt/abort/event callback remain intact. Agent now exposes mutex-protected Busy, Config, and SetModel so RPC/TUI code can inspect or mutate future provider labels without racing active prompts.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
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

### `(*Agent) Busy() bool`

Logic:

- Read the busy flag under the agent mutex.
- Return true while a prompt owns the active cancel function.

Acceptance:

- RPC and TUI can reject or queue work without racing prompt startup.

### `(*Agent) Config() Config`

Logic:

- Copy the provider id, api, and model labels under the agent mutex.
- Return no provider secrets.

Acceptance:

- headless state responses can report current model identity.

### `(*Agent) SetModel(providerID string, api string, model string) error`

Logic:

- Reject changes while busy.
- Trim and validate required labels.
- Update provider id, api, and model under the mutex.

Acceptance:

- RPC model commands can affect later provider requests;
- active streams keep their existing request labels.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050AbortStopsProviderAndTools`
- `TestAgentSetModelAffectsNextPrompt`
