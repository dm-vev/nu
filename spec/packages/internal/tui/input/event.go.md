# `internal/tui/input/event.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define decoded input events and bracketed paste markers.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define decoded input events and bracketed paste markers.

## Code Style

Decode just enough terminal protocol to hand complete sequences to editor/key handling. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Types And Constants

### `type Event struct {`

Logic:
- Event is one decoded terminal input sequence.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

### `const (...)`

Logic:
- Define bracketed paste start/end sequences preserved by the decoder.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

This file declares data/constants only; behavior is exercised through files in the same package.
