# `internal/tui/engine/lifecycle.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Start and stop terminal UI modes.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Start and stop terminal UI modes.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.
## Functions

### `func (t *TUI) Start() error`

Logic:
- Start initializes terminal state.
- Enable bracketed paste and hidden cursor state before rendering.
- Do not enable mouse reporting because it breaks normal terminal text selection/copy.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *TUI) Stop() error`

Logic:
- Stop restores terminal state and moves below the rendered frame.
- Disable any leftover mouse reporting, disable bracketed paste, and restore cursor visibility.

Acceptance:
- Terminal state is restored or cleanup is returned on every successful setup path.
