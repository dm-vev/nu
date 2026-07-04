# `internal/tui/components/footer/footer.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-style footer component for cwd, branch, context, provider, and model display.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Pi-style footer component for cwd, branch, context, provider, and model display.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Footer struct {`

Logic:
- Footer renders cwd, branch, context usage, provider, and model identity.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(opts Options) *Footer`

Logic:
- New creates a footer component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (f *Footer) SetOptions(opts Options)`

Logic:
- SetOptions replaces footer data.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (f *Footer) Options() Options`

Logic:
- Options returns the current footer data.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (f *Footer) Invalidate() {}`

Logic:
- Invalidate exists for the component interface.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
