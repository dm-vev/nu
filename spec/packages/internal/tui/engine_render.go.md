# `internal/tui/engine_render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Render a bottom-scrolled viewport of the component tree.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render editor/content/component lines for a fixed width.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.
## Functions

### `func (t *TUI) RenderNow() error`

Logic:
- RenderNow renders the current component tree.
- It allocates flexible filler rows before reset/padding, extracts cursor markers, pads to terminal height, clears on the first full render, and diffs same-size updates.
- When content exceeds terminal height, it keeps the bottom viewport by default so editor and footer remain visible.
- When `scrollOffset` is non-zero, it slices an older viewport and clamps the offset to available overflow.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.
- The first render clears the screen; same-size streaming updates do not clear again.
- Overflowing chat content autoscrolls to the bottom viewport.
- Manual scroll offset exposes older overflowing rows without terminal newline scrolling.

### `func (t *TUI) renderAnchored(width int, rows int) []string`

Logic:
- Render fixed children, count flexible fillers, and assign remaining terminal rows to fillers so components after a filler anchor to the bottom.

Acceptance:
- Editor/footer remain at the bottom when total fixed content is shorter than terminal height.
