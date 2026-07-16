# `internal/rpc/prompt.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Agent event forwarding and asynchronous prompt transitions were split from command dispatch.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Run asynchronous prompts, forward agent events, and record durable turn state.

## Code Style

Mark busy before goroutine start; never hold the state lock across provider work or protocol writes.

## Owned Logic

- `Emit` records final assistant text and serializes every agent event.
- `startPrompt` validates text/agent/idle state, drains steering, records the user message, and starts tracked work.
- `finishPrompt` clears busy state, emits non-cancellation errors, and starts queued follow-up work.
- Event helpers extract final text from typed or generic event data.

## Acceptance

- Prompt acceptance is immediate and correlated.
- Immediate commands observe the server as busy.
- Follow-ups run only after the current prompt becomes idle.

## Tests

- `TestNUF171RPCPromptResponseCorrelation`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
- `TestNUF052SteeringDeliveredBeforeNextProviderCall`
