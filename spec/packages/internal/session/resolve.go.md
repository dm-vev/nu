# `internal/session/resolve.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Path/full-ID/partial-ID and latest-by-cwd resolution were split from store operations.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Resolve session selectors and locate the latest session for a working directory.

## Code Style

Use header-only reads for discovery and return typed not-found/ambiguous errors.

## Owned Logic

- `Resolve` accepts existing JSONL paths, full IDs, or unambiguous ID prefixes.
- `LatestByCWD` compares cleaned header cwd and selects newest creation time, then ID.
- Helpers validate direct paths and enumerate root JSONL refs only.

## Acceptance

- Explicit files outside the store root can be resumed.
- Partial ID collisions return `ErrSessionAmbiguous`.
- Continue lookup is deterministic and ignores non-JSONL entries.

## Tests

- `TestNUF081ContinueLatestByCWD`
- `TestNUF081ResumeByPathOrPartialID`
