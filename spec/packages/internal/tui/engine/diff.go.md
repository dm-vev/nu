# `internal/tui/engine/diff.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Write full or differential synchronized terminal updates without newline scrolling near terminal bottom.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Write full or differential synchronized terminal updates.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func (t *TUI) fullRender(lines []string, width int, rows int, clear bool, cursor engineCursorPosition) error`

Logic:
- Write every line in one synchronized output buffer and reset previous render state.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.
- The first render writes the whole component tree without `CSI 2J`.

### `func (t *TUI) diffRender(lines []string, width int, rows int, cursor engineCursorPosition) error`

Logic:
- Find changed range, position to each changed row absolutely, clear changed
  lines, write replacements, and update cache.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.
- Ordinary assistant streaming updates clear and rewrite only changed lines.
- Editor updates near the bottom do not use newline-based redraws that can
  scroll the terminal.

### `func engineChangedRange(oldLines []string, newLines []string) (int, int)`

Logic:
- Return first/last changed line indexes across old and new line slices.

Acceptance:
- Returns `(-1, -1)` when no line differs.
