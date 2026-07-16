# `internal/tui/input/decoder_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Raw terminal input decoder that groups escape sequences and bracketed paste.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Raw terminal input decoder that groups escape sequences and bracketed paste.

## Code Style

Decode just enough terminal protocol to hand complete sequences to editor/key handling. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestInputDecoderReadsPrintableUTF8AndEscapeSequence(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestInputDecoderRewrapsBracketedPaste(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.
