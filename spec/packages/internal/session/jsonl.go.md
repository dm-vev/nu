# `internal/session/jsonl.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Session JSONL read/parse/write helpers were split from store operations.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Decode, validate, and encode session headers and entries using JSONL framing.

## Code Style

Use bounded Scanner buffers, contextual path errors, and entry marshal helpers.

## Owned Logic

- `readSession` validates the first header line and decodes all following entries.
- `readHeader` reads only the first line for cheap discovery.
- `parseSession` validates imported bytes, including the complete entry tree.
- `writeJSONLine` writes one marshaled value followed by LF.

## Acceptance

- Empty, malformed, or wrong-schema headers fail clearly.
- Import parsing rejects broken parent links before disk writes.
- Header-only discovery does not read whole session files.

## Tests

- `TestNUF080SessionLoadRejectsBrokenParent`
- `TestSessionExportImportRoundTrip`
- `TestSessionImportRejectsOversizedInput`
