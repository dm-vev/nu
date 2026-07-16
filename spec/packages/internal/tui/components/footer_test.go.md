# `internal/tui/components/footer_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-style footer component for cwd, branch, context, provider, and model display.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Pi-style footer component for cwd, branch, context, provider, and model display.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestFooterFooterFormatsPathAndStats(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.
