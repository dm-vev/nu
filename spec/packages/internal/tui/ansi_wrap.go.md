# `internal/tui/ansi_wrap.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Wrap ANSI text by visible width while preserving escape sequences.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Wrap ANSI text by visible width while preserving escape sequences.

## Code Style

Keep helpers pure, allocation-light, and independent from component state. Never count escape bytes as visible cells. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.

## Functions

### `func ANSIWrapText(text string, width int) []string`

Logic:
- WrapText wraps text to width while preserving ANSI sequences.

Acceptance:
- Returned lines preserve escape sequences and are suitable for final `ANSIPadRight` width enforcement.

### `func ansiWrapParagraph(text string, width int) []string`

Logic:
- Wrap one paragraph by visible width, carrying ANSI sequences into output without counting them.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func ansiSplitAtLastSpace(text string) (string, string)`

Logic:
- Split a candidate line at the last whitespace so word wrapping keeps the trailing word for the next line.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
