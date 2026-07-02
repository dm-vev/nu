# `internal/agent/queue.go`

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

Steering and follow-up message queues.

## Code Style

Small mutex-protected struct. No goroutines.

## Functions

### `(*Queue) AddSteering(msg Prompt)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Append steering message and emits full queue snapshot.

Acceptance:

- appends steering message and emits full queue snapshot.

### `(*Queue) NextSteering(mode DeliveryMode) []Prompt`

Logic:

- Lock the queue and snapshot the current steering messages.
- For single-delivery mode, remove and return only the oldest steering prompt.
- For batch mode, remove and return all queued steering prompts in insertion order.
- Emit a queue update after removal so TUI/RPC state remains consistent.

Acceptance:

- returns one or all queued messages according to mode;
- removes returned messages.

### `(*Queue) AddFollowUp(msg Prompt)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Follow-up remains queued until agent idle.

Acceptance:

- follow-up remains queued until agent idle.

Tests:

- `TestNUF052SteeringDeliveredBeforeNextProviderCall`
- `TestNUF052FollowUpWaitsForIdle`
