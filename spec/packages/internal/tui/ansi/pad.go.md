# `internal/tui/ansi/pad.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pad/truncate visible cells to exact terminal width.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Pad/truncate visible cells to exact terminal width.

## Code Style

Keep helpers pure, allocation-light, and independent from component state. Never count escape bytes as visible cells. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.
## Functions

### `func PadRight(text string, width int) string`

Logic:
- PadRight pads visible cells to width.

Acceptance:
- Tests cover ANSI-preserving visible-width behavior.
