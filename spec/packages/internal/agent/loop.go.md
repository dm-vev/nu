# `internal/agent/loop.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Text-only provider loop exists; tool-call continuation and message persistence are deferred until their specs are reopened.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Text-only provider turn loop. Consumes `spec/protocols/provider-stream.md`.

## Code Style

Keep the loop readable. Do not hide state transitions behind clever channels.

## Functions

### `runTurn(ctx context.Context, state *State, input TurnInput) error`

Logic:

- Build a provider request from the current user prompt and selected provider/model fields.
- Emit `turn_start` before calling provider.
- Start provider stream with the same `ctx` used for abort.
- For every provider event, call `handleProviderEvent` and emit corresponding
  message/update events.
- On normalized provider `error`, return `provider.ErrStream` with context.
- On `done`, emit final `turn_end` with accumulated text.
- Stop on context cancellation, abort, premature stream close, or validation error.

Acceptance:

- sends context to provider;
- accumulates assistant stream into messages;
- stops on final provider done, abort, provider error, or unrecoverable error.

### `handleProviderEvent(state *State, ev provider.Event) error`

Logic:

- Switch only on normalized provider event type.
- Emit `message_start` on provider start.
- Append `text_delta` to the accumulated response text and emit a live update.
- Emit `message_end` on provider done.
- Leave provider error handling to `runTurn`.
- Return an error for unknown normalized events.

Acceptance:

- converts start/text/done provider events to message events without provider-specific logic.

Tests:

- `TestNUF050TextOnlyTurnEnds`
- `TestNUF050AbortStopsProviderAndTools`
