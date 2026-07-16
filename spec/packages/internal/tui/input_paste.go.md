# `internal/tui/input_paste.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Read bracketed paste payloads to completion.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Read bracketed paste payloads to completion.

## Code Style

Decode just enough terminal protocol to hand complete sequences to editor/key handling. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.

## Functions

### `func (d *Decoder) readPaste() (string, error)`

Logic:
- Read until the bracketed paste terminator and return literal paste content.

Acceptance:
- The returned string excludes the end marker and preserves pasted newlines.

### `func inputHasPasteEnd(buffer []byte) bool`

Logic:
- Check whether the byte buffer currently ends with the bracketed paste terminator.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
