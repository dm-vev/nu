# `internal/extension/protocol.go`

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

JSONL process-extension protocol definitions.
Implements `spec/protocols/extension-jsonl.md`.

## Code Style

Protocol structs are versioned and golden-tested. Keep stdout protocol separate
from stderr diagnostics.

## Types

### `Frame`, `Request`, `Response`, `Event`

Logic:

- Represent handshake, registration, hook, UI, and shutdown frames from
  `spec/protocols/extension-jsonl.md`.
- Keep request, response, and event frames distinct after decoding.
- Preserve frame id as opaque string for correlation.
- Include protocol version on handshake frames only.

Acceptance:

- every request can carry id for correlation;
- supports register tool/command/keybinding/flag, hook results, UI requests,
  and shutdown.
- follows handshake and capability rules in `spec/protocols/extension-jsonl.md`.

## Functions

### `DecodeFrame(line []byte) (Frame, error)`

Logic:

- Trim one trailing CR before JSON decode.
- Decode `type` first.
- Dispatch to typed frame payload.
- Reject frames that require `id` but omit it.
- Return typed protocol error without killing host directly.

Acceptance:

- supports every frame in `spec/protocols/extension-jsonl.md`;
- preserves unknown extension-owned payload details only where allowed.

### `EncodeFrame(frame Frame) ([]byte, error)`

Logic:

- Validate frame type/id/capabilities before marshaling.
- Marshal compact JSON without trailing newline.

Acceptance:

- produces one JSON object per frame;
- never includes stderr diagnostics.

Tests:

- `TestExtensionProtocolRoundTrip`
