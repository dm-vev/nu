# `internal/tui/editor/editor_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Focused multiline input editor with rune-safe mutation and bordered rendering.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Focused multiline input editor with rune-safe mutation and bordered rendering.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestEditorEditorWrapsInput(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestEditorEditorSubmit(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestEditorEditorCanUseASCIIBorder(t *testing.T)`

Logic:
- Configure the editor border rune to `-` and assert both border rows use only ASCII hyphens.

Acceptance:
- The test fails if limited terminals cannot avoid Unicode prompt borders.

### `func TestEditorEditorHandlesUnicodeCursorAndPaste(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestEditorEditorForwardDeleteUsesRuneCursor(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.
