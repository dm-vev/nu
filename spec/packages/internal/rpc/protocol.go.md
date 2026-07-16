# `internal/rpc/protocol.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Input framing, command envelopes, line parsing, and response construction moved from `jsonl.go`/`server.go`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Define RPC wire records and parse strict LF-delimited JSONL commands.

## Code Style

Use `bufio.Reader`, not `Scanner`; preserve command IDs and return structured parse failures.

## Owned Logic

- `rpcMessage`, `commandEnvelope`, and `response` define internal wire shapes, including accepted snake/camel aliases.
- `ReadLines` splits only on LF, accepts one CR before LF, emits a final unterminated line, and propagates callback errors.
- `handleLine` decodes, ignores extension UI responses, dispatches, and writes one response when requested.
- `success` and `failure` build correlated response objects.

## Acceptance

- Large lines have no Scanner token ceiling.
- Invalid JSON yields a structured parse response.
- IDs and command names survive response construction.

## Tests

- `TestRPCJSONLReadLinesStrictLF`
- `TestRPCJSONLReadLinesReturnsCallbackError`
- `TestNUF171RPCPromptResponseCorrelation`
