# `internal/tui/box_box.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Padding/background container component.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Padding/background container component.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Box struct {`

Logic:
- Box applies padding and background around child components.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func NewBox(opts BoxOptions) *Box`

Logic:
- New creates a box.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
