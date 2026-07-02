# `internal/rpc/protocol.go`

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

RPC JSONL command/response/event schema.
Implements `spec/protocols/rpc-jsonl.md`.

## Code Style

Protocol is line-delimited JSON only. No generic line splitting beyond LF.

## Types

### `Command`, `Response`

Logic:

- Represent commands and responses exactly as `spec/protocols/rpc-jsonl.md`.
- Keep command payloads as typed structs, not generic maps, after decoding.
- Preserve client request id as an opaque string.
- Include command name, success flag, error object, and data object in every
  command response.

Acceptance:

- supports prompt, steer, follow_up, abort, new_session, state, settings, and
  shutdown;
- response includes request id when provided.
- stdout-compatible JSON shape matches `spec/protocols/rpc-jsonl.md`.

## Functions

### `DecodeCommand(line []byte) (Command, error)`

Logic:

- Trim one trailing CR before JSON decode.
- Reject empty lines.
- Decode common fields first, then decode typed payload by `type`.
- Return typed protocol error with request id when available.

Acceptance:

- splits only on caller-provided LF frame;
- returns typed command payloads;
- preserves request id on malformed typed payload.

### `EncodeResponse(resp Response) ([]byte, error)`

Logic:

- Validate response has command and success/error consistency.
- Marshal compact JSON without trailing newline.
- Preserve request id exactly when present.

Acceptance:

- produces one JSON object suitable for stdout JSONL;
- never writes diagnostics text.

Tests:

- `TestRPCProtocolRoundTrip`
