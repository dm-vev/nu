# `internal/tools/coding/filesystem.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Coding mutation and temporary-output helpers live in `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Own the shared filesystem mutation lock and persisted full bash output.

## Code Style

Keep the global lock until measured throughput requires per-path locks; wrap temp-file errors.

## Owned Logic

- `mutationMu` serializes write/edit mutations within the process.
- `persistTempOutput` writes complete truncated command output to a named temp log.

## Acceptance

- Concurrent mutations do not produce partial same-process writes.
- Truncated bash output is recoverable from the returned path.

## Tests

- `TestNUF071ConcurrentWritesSamePathSerialize`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
