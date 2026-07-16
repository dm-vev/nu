# `internal/rpc/server.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Server lifecycle and shared state remain here; protocol, dispatch, prompt, session, queue, state, bash, and writing moved to cohesive files.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Own construction, lifecycle, cancellation, and shared state for one JSONL RPC server.

## Code Style

Keep RPC independent of `app`; protect mutable state and stdout with separate mutexes.

## Owned Logic

- `Options` and `Server` define injectable IO, model/session state, queues, messages, agent, locks, and prompt wait tracking.
- `NewServer` normalizes IO and seeds session/model/queue defaults without side effects.
- `SetAgent`, `Run`, `waitIdle`, abort/shutdown helpers, and ID/string helpers own server lifecycle.

## Acceptance

- EOF, shutdown, context cancellation, and write failure terminate predictably.
- Shutdown aborts active work and waits for prompt goroutines.
- Construction requires no provider or process global.

## Tests

- `TestNUF171RPCShutdownWritesFinalResponse`
- `TestNUF171RPCPromptResponseCorrelation`
