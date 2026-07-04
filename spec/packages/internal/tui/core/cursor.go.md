# `internal/tui/core/cursor.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define cursor marker and extracted cursor position data.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define cursor marker and extracted cursor position data.

## Code Style

Keep interfaces tiny and structural. Containers own ordering only, not layout policy. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Types And Constants

### `type Cursor struct {`

Logic:
- Cursor stores a logical cursor position.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

### `const CursorMarker = "\x1b_pi:c\x07"`

Logic:
- CursorMarker is a zero-width marker stripped by the engine before writing.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

This file declares data/constants only; behavior is exercised through files in the same package.
