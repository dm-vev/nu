# `internal/session/store.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Core store append/load/write behavior remains here; types, JSONL, resolution, branches, transfer, and compaction moved to dedicated session files.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Open a session store and append, load, validate, or create complete session files.

## Code Style

Use stdlib filesystem calls and one in-process lock around append validation plus write.

## Owned Logic

- `OpenStore` cleans the root without filesystem side effects.
- `Append` validates entries and parent/duplicate constraints before appending, creating a schema header for new sessions.
- `Load` reads JSONL and reconstructs the tree.
- Path and write helpers honor explicit paths, clean optional cwd, validate trees, and create targets exclusively.

## Acceptance

- Invalid parents or duplicate IDs are never appended.
- New files preserve ref cwd and use private permissions.
- Loaded sessions contain validated entries and tree state.

## Tests

- `TestNUF080SessionAppendBuildsTree`
- `TestNUF080SessionLoadRejectsBrokenParent`
- `TestNUF080SessionAppendRejectsBrokenParent`
- `TestSessionAppendRejectsDuplicateID`
- `TestSessionAppendUsesRefCWD`
