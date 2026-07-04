# `internal/tui/components/header/header.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-style startup header with compact and expanded help.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Pi-style startup header with compact and expanded help.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Header struct {`

Logic:
- Header renders Pi-style startup hints and onboarding text.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(opts Options) *Header`

Logic:
- New creates a header component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (h *Header) SetExpanded(expanded bool)`

Logic:
- SetExpanded switches between compact and full help.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (h *Header) Toggle()`

Logic:
- Toggle flips the expansion state.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (h *Header) Expanded() bool`

Logic:
- Expanded reports the current expansion state.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (h *Header) Invalidate() {}`

Logic:
- Invalidate exists for the component interface.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
