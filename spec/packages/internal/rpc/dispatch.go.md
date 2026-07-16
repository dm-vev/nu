# `internal/rpc/dispatch.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Pi-compatible command routing was split from server lifecycle.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Route each RPC command to the state, prompt, queue, session, model, bash, or shutdown owner.

## Code Style

Keep the switch explicit so protocol coverage is auditable; return structured failures for invalid or unknown commands.

## Owned Logic

- `handleCommand` recognizes the `NUF-171` command set.
- Busy prompts are rejected unless requested as steering or follow-up work.
- Command aliases are normalized before mutation and every synchronous command returns one correlated response.

## Acceptance

- Pi command names are recognized without provider calls for state-only commands.
- Missing model/session names and unknown commands fail clearly.
- Shutdown responds before server termination.

## Tests

- `TestNUF171RPCRecognizesPiCommandSet`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
- `TestNUF171RPCShutdownWritesFinalResponse`
