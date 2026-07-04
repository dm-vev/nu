# `internal/agent/agent.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 5d9629b
Implementation Comments: Prompt/abort/event callback remain intact. Agent exposes mutex-protected Busy, Config, SetModel, SetProviderModel, Reset, remembers successful turn history, and forwards provider tool definitions into requests.

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
- Preserve provider-facing tool definitions for request construction.
- Start no provider call until `Prompt`.

Acceptance:

- initializes provider settings, event callback, tools map, and tool definitions;
- starts no provider call until `Prompt`.

### `(*Agent) Prompt(ctx context.Context, input Prompt) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Reject concurrent prompt without queue behavior.
- Start one provider turn from remembered successful history plus the supplied prompt text.
- Store the completed turn messages after success so later prompts can answer questions about prior user requests and assistant work.
- Emit agent/turn/message events through the callback.

Acceptance:

- rejects concurrent prompt without queue behavior;
- sends one prompt to provider;
- emits agent/turn/message events.
- forwards tool definitions to the provider request.
- `TestAgentPromptIncludesPreviousTurns` fails if the second prompt loses prior user/assistant context.

### `(*Agent) Reset()`

Logic:

- Clear remembered prompt history under the agent mutex.

Acceptance:

- `/new` can start a fresh chat instead of leaking prior context into the next provider request.

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

- Delegate to `SetProviderModel` without changing the existing streamer.

Acceptance:

- RPC model commands can affect later provider requests;
- active streams keep their existing request labels.

### `(*Agent) SetProviderModel(streamer provider.Streamer, providerID string, api string, model string) error`

Logic:

- Reject changes while busy.
- Trim and validate required labels.
- Replace the provider streamer when a non-nil streamer is supplied.
- Update provider id, api, and model under the mutex.

Acceptance:

- TUI model selection can switch both the displayed model and the provider client used by later prompts.
- `TestAgentSetProviderModelSwitchesStreamer` fails if prompts still use the previous streamer.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestAgentPromptIncludesPreviousTurns`
- `TestNUF050ProviderRequestIncludesToolDefinitions`
- `TestNUF050ToolCallFeedsResultBackToProvider`
- `TestNUF050AbortStopsProviderAndTools`
- `TestAgentSetModelAffectsNextPrompt`
- `TestAgentSetProviderModelSwitchesStreamer`
