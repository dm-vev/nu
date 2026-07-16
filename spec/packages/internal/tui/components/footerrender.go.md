# `internal/tui/components/footerrender.go`

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

### `func (f *Footer) Render(width int) []string`

Logic:
- Render produces cwd and stats/model lines.
- Render the optional notice on the right side of the cwd/path line.
- Use the current `Used` and `Context` option values for the left stats segment.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.

### `func footerAlignStats(left string, right string, width int) string`

Logic:
- Fit left and right text into one terminal-width line, truncating only when required.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func footerStylePathLine(line string, opts FooterOptions) string`

Logic:
- Dim the path segment and apply notice styling only to the right-side notice.

Acceptance:
- `TestTUIRateLimitShowsFooterNotice` fails if `Rate limit` is missing or not red.
