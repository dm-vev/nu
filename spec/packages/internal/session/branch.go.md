# `internal/session/branch.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Fork/clone target creation was split from store operations.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Create new session files from a selected source branch path.

## Code Style

Reuse tree/path and store-write primitives; do not rewrite source sessions.

## Owned Logic

- `Fork` copies the root-to-requested-entry path.
- `Clone` copies the root-to-active-leaf path.
- `createFromEntries` checks cancellation/target ID, assigns a fresh header ID/time, and writes exclusively.

## Acceptance

- Fork and clone preserve valid parent links and source metadata.
- Existing target files are not overwritten.
- Fork includes exactly the selected branch prefix; clone includes only the active branch.

## Tests

- `TestNUF081ForkStartsNewFileFromUserEntry`
- `TestNUF081CloneCopiesActiveBranch`
