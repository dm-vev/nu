# `internal/rpc/server.go`

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

Run Nu as a JSONL RPC server over stdin/stdout.
Uses `spec/protocols/rpc-jsonl.md`.

## Code Style

Stdout is protocol-only. Diagnostics go to stderr.

## Functions

### `Serve(ctx context.Context, opts ServerOptions) error`

Logic:

- Start one goroutine that subscribes to agent/session events and encodes them
  to stdout JSONL under a write mutex.
- Read stdin with LF-only framing; trim optional CR.
- For each command, decode with `rpc.DecodeCommand`.
- Dispatch prompt/steer/follow_up/abort/new_session/state/settings/shutdown to
  services.
- Write exactly one response for every decoded command.
- For malformed JSON without id, emit uncorrelated protocol error event.
- Never write human diagnostics to stdout.
- On shutdown command, write response, cancel server context, drain owned
  goroutines, and return.

Acceptance:

- reads LF-delimited JSON commands;
- dispatches to agent/session services;
- streams events concurrently;
- rejects prompt during active stream unless behavior is specified.

Tests:

- `TestNUF171RPCPromptResponseCorrelation`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
