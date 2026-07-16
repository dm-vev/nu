# `internal/tui/components/footerformat.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Format footer path, token counts, and aligned model labels.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Format footer path, token counts, and aligned model labels.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func FormatPath(cwd string, home string, branch string) string`

Logic:
- FormatPath returns cwd shortened relative to home and annotated with branch.

Acceptance:
- Paths inside home render as `~` or `~/relative`; paths outside home stay absolute.

### `func FormatTokens(count int) string`

Logic:
- FormatTokens returns a compact context token count.

Acceptance:
- Counts below 1000 stay exact; larger values render with `k` or `M` suffixes.

### `func footerStatsLeft(used int, contextWindow int) string`

Logic:
- Build the left footer stats segment as used-context percent/window plus auto-compaction marker.
- Derive the percentage from `used/contextWindow` and keep one decimal place.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func footerModelRight(provider string, model string) string`

Logic:
- Join provider and model display label with `/`, trimming empty sides.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func footerFirstNonEmpty(values ...string) string`

Logic:
- Return the first trimmed non-empty candidate or an empty string.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
