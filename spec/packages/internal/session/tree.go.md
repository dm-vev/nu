# `internal/session/tree.go`

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

In-memory session tree operations.

## Code Style

Pure functions over loaded entries. No filesystem.

## Functions

### `BuildTree(entries []Entry) (*Tree, error)`

Logic:

- Iterate entries in file order.
- Reject duplicate ids immediately.
- For each non-root entry, require parent id to already exist; this enforces
  append-only parent-before-child semantics.
- Build parent, child, depth, and leaf indexes.
- Identify roots and leaves.
- Set active leaf using persisted state entry when valid, otherwise last entry.

Acceptance:

- detects duplicate ids;
- detects missing parents;
- computes roots, children, leaves, and active leaf.

### `PathTo(tree *Tree, leaf string) ([]Entry, error)`

Logic:

- Look up leaf id.
- Walk parent links to root, collecting entries.
- Reverse collected entries.
- Return error if a cycle is detected despite previous validation.

Acceptance:

- returns ordered root-to-leaf path.

### `MoveLeaf(tree *Tree, entryID string) error`

Logic:

- Verify entry id exists.
- Set in-memory active leaf to that id.
- Do not append or rewrite session data; persistence, if desired, is a separate
  session-state entry.

Acceptance:

- moves active leaf without changing entries.

Tests:

- `TestSessionBuildTreeRejectsDuplicateID`
- `TestNUF082SelectingAssistantEntryMovesLeaf`
