# `internal/tui/engine_engine.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define and construct the renderer engine.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define and construct the renderer engine.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type TUI struct {`

Logic:
- TUI manages a component tree and terminal diff rendering.
- It tracks the current scroll offset from the bottom viewport.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(term *Terminal, opts Options) *TUI`

Logic:
- New creates a TUI engine.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *TUI) Terminal() *Terminal`

Logic:
- Terminal returns the underlying terminal.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (t *TUI) ScrollBy(delta int) bool`

Logic:
- Move the viewport away from or toward the bottom by `delta` rows.
- Clamp negative offsets to zero; final upper clamp happens during render when content height is known.

Acceptance:
- `TestEngineEngineScrollsOverflowingViewport` fails if positive scroll no longer exposes older rows.

### `func (t *TUI) ScrollToBottom() bool`

Logic:
- Reset manual scrolling and resume bottom-following behavior.

Acceptance:
- Raw End key handling can restore the live bottom viewport.
