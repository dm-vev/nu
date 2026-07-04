# `internal/tui/core/container_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Minimal component primitives shared by all TUI packages.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui/...`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Minimal component primitives shared by all TUI packages.

## Code Style

Keep interfaces tiny and structural. Containers own ordering only, not layout policy. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.

## Types And Constants

### `type staticComponent []string`

Logic:
- Provide a minimal component fixture that returns static lines for container-order tests.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func (s staticComponent) Render(width int) []string`

Logic:
- Render the component state to width-bounded terminal lines without side effects outside cache/state.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.

### `func TestContainerRendersChildrenInOrder(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.
