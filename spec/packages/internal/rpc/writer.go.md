# `internal/rpc/writer.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: JSONL output and first-error tracking moved from `jsonl.go`/`server.go` to the writer owner.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Serialize RPC responses/events as strict JSONL and retain the first stdout failure.

## Code Style

Marshal before writing, append exactly one LF, and serialize all writes with `writeMu`.

## Owned Logic

- `WriteLine` emits one JSON object plus LF with contextual errors.
- `write` serializes stdout access and refuses writes after the first failure.
- `setWriteErr` and `currentWriteErr` safely preserve/read the first error.

## Acceptance

- Concurrent events and responses cannot interleave bytes.
- Every successful write is one LF-delimited JSON value.
- The server returns the first protocol output error.

## Tests

- `TestRPCJSONLWriteLine`
- `TestNUF171RPCPromptResponseCorrelation`
