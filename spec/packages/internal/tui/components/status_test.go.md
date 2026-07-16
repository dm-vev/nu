# `internal/tui/components/status_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Transient status-line component.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Transient status-line component.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestStatusStatusAlwaysReservesOneLine(t *testing.T)`

Logic:
- Verify idle and busy status both render exactly one row.

Acceptance:
- The test fails if idle status collapses the layout row.

### `func TestStatusStatusStepAnimatesLabel(t *testing.T)`

Logic:
- Verify `Step` changes the rendered busy status label.

Acceptance:
- The test fails if the working animation no longer advances.

### `func TestStatusStatusUsesClaudeLikeFrames(t *testing.T)`

Logic:
- Verify the first status frames use the Claude-like interpunct and asterisk glyphs.

Acceptance:
- The test fails if the spinner regresses to dot suffixes.

### `func TestStatusStatusCanUseASCIIFrames(t *testing.T)`

Logic:
- Verify an injected frame set renders the ASCII spinner sequence.

Acceptance:
- The test fails if limited terminals cannot replace Unicode spinner glyphs.
