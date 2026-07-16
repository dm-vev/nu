# `internal/tools/coding/ls.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Directory-list behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

List one cwd-contained directory as stable JSON.

## Code Style

Use `os.ReadDir`, include dotfiles, sort names, and suffix directories with `/`.

## Owned Logic

- `RunLS` defaults path to `.`, validates containment/context/directory type, reads and sorts entries, and returns a bounded list.

## Acceptance

- Output is sorted, includes dotfiles, and marks directories.
- Non-directories and lexical/symlink cwd escapes fail.

## Tests

- `TestNUF076LsSortedWithDirs`
- `TestNUF076LsRejectsNonDirectory`
- `TestLsRejectsPathEscape`
