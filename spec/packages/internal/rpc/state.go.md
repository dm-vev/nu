# `internal/rpc/state.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: RPC state projections and message/tree queries were split from command dispatch.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Build immutable protocol views of current server, queue, model, session, and message state.

## Code Style

Copy slices/maps before returning and hold the state mutex while reading server fields.

## Owned Logic

- Build state/settings/model/queue/session-stat responses with Pi-compatible aliases.
- Select user messages, entries after an ID, a linear tree, active leaf, and last assistant text.
- Clone map and message data so callers cannot mutate server state.

## Acceptance

- State can be queried while idle without a provider call.
- Returned queues/messages/settings do not alias mutable server slices/maps.
- Tree and leaf IDs reflect current in-memory message order.

## Tests

- `TestNUF171RPCRecognizesPiCommandSet`
