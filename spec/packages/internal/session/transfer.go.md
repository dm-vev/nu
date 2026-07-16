# `internal/session/transfer.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Session export/import and bounded input handling were split from store operations.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Export validated session bytes and import bounded, validated JSONL into a new target.

## Code Style

Validate before writing and bound user-controlled input before JSON parsing.

## Owned Logic

- `Export` loads for validation, then writes the stored bytes unchanged.
- `Import` reads at most the configured limit plus one byte, parses header/tree, applies an optional target ID, and creates the target exclusively.

## Acceptance

- Export/import round-trips valid sessions.
- Oversized or invalid imports create no file.
- Existing targets are never overwritten.

## Tests

- `TestSessionExportImportRoundTrip`
- `TestSessionImportRejectsOversizedInput`
