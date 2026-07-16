# `internal/tui/text_text.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Wrapping, cached text component with optional background fill.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Wrapping, cached text component with optional background fill.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Text struct {`

Logic:
- Text displays wrapped multi-line text.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func NewText(value string, opts TextOptions) *Text`

Logic:
- New creates a text component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *Text) SetText(value string)`

Logic:
- SetText changes text and clears cached lines.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *Text) Text() string`

Logic:
- Text returns the current text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
