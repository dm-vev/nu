# `internal/tui/components/boxbackground.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Apply component background fill to width-padded lines.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Apply component background fill to width-padded lines.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func (b *Box) applyBackground(line string, width int) string`

Logic:
- Pad/truncate a line to width and apply the optional background callback to the whole padded line.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
