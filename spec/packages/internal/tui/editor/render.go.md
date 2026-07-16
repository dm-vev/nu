# `internal/tui/editor/render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Render editor lines without moving the cursor to a new visual row for trailing spaces.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render editor/content/component lines for a fixed width.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func (e *Editor) Render(width int) []string`

Logic:
- Insert the cursor marker into the logical line before wrapping.
- Wrap and pad editor content between two border lines.
- Use the configured border rune, defaulting to the Unicode horizontal rule when no override is set.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.
- A trailing space at the cursor does not add an extra visual editor line.
- ASCII mode renders `-` border lines instead of Unicode horizontal rules.

### `func editorInsertMarker(text string, col int) string`

Logic:
- Insert `CursorMarker` at a rune column after clamping the column to the rendered text length.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
