# `internal/app/jsonsession.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: JSON-mode session framing and ID generation were split from runtime composition.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Create JSON-mode session headers, random session IDs, and single-line JSON records.

## Code Style

Use stdlib crypto randomness and JSON encoding; wrap generation and write errors.

## Owned Logic

- `jsonSessionHeader` owns the JSON session header fields.
- `newJSONSessionHeader` uses an injected ID or a random UUID-shaped ID and defaults the app version to `dev`.
- `newSessionID` sets UUID version/variant bits from 16 random bytes.
- `writeJSONLine` marshals one value and appends exactly one LF.

## Acceptance

- Headers identify schema 1, Nu, cwd, version, and a stable invocation ID.
- Every encoded value occupies one JSONL record.

## Tests

- `TestJSONSessionHeaderDefaults`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
