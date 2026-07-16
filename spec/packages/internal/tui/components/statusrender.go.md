# `internal/tui/components/statusrender.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Render editor/content/component lines for a fixed width.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render editor/content/component lines for a fixed width.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.
## Functions

### `func (s *Status) Render(width int) []string`

Logic:
- Always render exactly one terminal row so the editor remains anchored below the status line.
- Idle state renders a blank padded row.
- Busy state renders the current configured spinner frame plus the styled label.
- Alert state renders the same spinner through a yellow-to-red 256-color gradient.

Acceptance:
- ANSI-stripped output never exceeds the requested width.
- `TestStatusStatusAlwaysReservesOneLine` fails if idle status stops reserving the row.
- `TestStatusStatusCanUseASCIIFrames` fails if configured ASCII frames are ignored.

### `func statusAlertStyle(value string, frame int) string`

Logic:
- Pick a deterministic 256-color foreground from yellow toward red based on frame index.

Acceptance:
- Alert animation remains one row and changes color without changing layout.
