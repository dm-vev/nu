# `internal/tui/components/textcache.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Cache rendered text lines by source text and width.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Cache rendered text lines by source text and width.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type textCache struct {`

Logic:
- Store the last rendered source text, width, and line slice for reuse.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func (t *Text) cached(width int) ([]string, bool)`

Logic:
- Return cached lines only when both source text and width match the current component state.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *Text) store(width int, lines []string)`

Logic:
- Store the rendered line slice with the source text and width that produced it.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *Text) Invalidate()`

Logic:
- Invalidate clears cached lines.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
