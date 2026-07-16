# `internal/tools/coding/write.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Write behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Create or overwrite one cwd-contained file.

## Code Style

Use stdlib filesystem calls and the shared mutation lock.

## Owned Logic

- `RunWrite` decodes path/content, validates containment/context, serializes mutation, creates parent directories, writes mode `0644`, and reports path/byte count.

## Acceptance

- Nested files are created and existing files replaced without partial same-process writes.
- Lexical and symlink-parent cwd escapes fail.

## Tests

- `TestNUF071WriteCreatesFile`
- `TestNUF071ConcurrentWritesSamePathSerialize`
- `TestWriteRejectsPathEscape`
- `TestWriteRejectsSymlinkParentEscape`
