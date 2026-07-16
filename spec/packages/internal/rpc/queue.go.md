# `internal/rpc/queue.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Steering and follow-up queue policy was split from prompt execution.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Queue, merge, and drain steering and follow-up prompt text.

## Code Style

Ignore blank queue entries and mutate queues only under the server mutex.

## Owned Logic

- Enqueue steering/follow-up text.
- Drain all queued items or one item according to `all`/`one-at-a-time` mode.
- Merge steering before prompt text and normalize unknown queue modes to `all`.

## Acceptance

- Steering is delivered before the next provider call.
- One-at-a-time mode preserves remaining queue order.
- Blank messages never enter a queue.

## Tests

- `TestNUF052SteeringDeliveredBeforeNextProviderCall`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
