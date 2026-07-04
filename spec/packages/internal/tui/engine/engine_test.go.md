# `internal/tui/engine/engine_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Component tree renderer and synchronized terminal diff writer.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui/...`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Component tree renderer and synchronized terminal diff writer.

## Code Style

Write one synchronized buffer per render. Prefer full render only for first paint or resize. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestEngineRendersComponentTreeAndDiffs(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestEngineRendersBottomViewportWhenContentOverflows(t *testing.T)`

Logic:
- Render more component lines than terminal rows.

Acceptance:
- The rendered frame contains the bottom content and not the first scrolled-off line.

### `func TestEngineDiffUsesAbsoluteRowsNearBottom(t *testing.T)`

Logic:
- Update a bottom-row component after an initial render.

Acceptance:
- Diff rendering uses absolute row positioning instead of newline-based redraw.
