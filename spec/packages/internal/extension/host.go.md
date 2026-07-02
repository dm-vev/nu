# `internal/extension/host.go`

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

Start, supervise, and stop extension processes.
Uses `spec/protocols/extension-jsonl.md`.

## Code Style

Use context cancellation and process cleanup. Never let extension stdout mix
with user stdout.

## Functions

### `StartHost(ctx context.Context, spec Spec, io HostIO) (*Host, error)`

Logic:

- Validate extension spec and trust decision before process start.
- Start process with stdin/stdout pipes reserved for protocol and stderr routed
  to diagnostics.
- Read first frame with handshake timeout.
- Validate protocol version, extension id, and requested capabilities.
- Send `hello_ack` with session/mode metadata.
- Start read loop that decodes frames and routes registrations, hook responses,
  and UI responses to registry/dispatch waiters.
- Stop host if stdout closes unexpectedly, protocol decode fails fatally, or
  context is cancelled.

Acceptance:

- starts configured process;
- validates handshake;
- reads JSONL frames until shutdown or failure.

### `(*Host) Close(ctx context.Context) error`

Logic:

- Atomically mark host closing so shutdown runs once.
- Send shutdown request if process is still alive and protocol writer works.
- Wait for response or timeout.
- Close stdin to signal no more requests.
- Kill process after timeout.
- Join read loop and return combined shutdown errors.

Acceptance:

- sends shutdown once;
- kills process after timeout.

Tests:

- `TestNUF160ExtensionShutdownRunsOnce`
