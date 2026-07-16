# `internal/tui/input/decoder.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Read raw bytes and group them into input events.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Read raw bytes and group them into input events.

## Code Style

Decode just enough terminal protocol to hand complete sequences to editor/key handling. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.

## Types And Constants

### `type Decoder struct {`

Logic:
- Decoder turns a raw byte stream into terminal key/input events.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(reader io.Reader) *Decoder`

Logic:
- New creates a decoder over reader.

Acceptance:
- Does not read from the stream during construction.

### `func (d *Decoder) Read() (Event, error)`

Logic:
- Read returns the next decoded event.

Acceptance:
- EOF is returned unchanged so callers can exit cleanly.

### `func (d *Decoder) readUTF8(first byte) (Event, error)`

Logic:
- Read enough bytes for a complete UTF-8 rune.

Acceptance:
- Invalid short reads return a wrapped `read utf8 input` error.

### `func (d *Decoder) readEscape() (Event, error)`

Logic:
- Read a complete escape sequence and delegate bracketed paste payload reads.

Acceptance:
- Bracketed paste is returned as one event with start/end markers preserved.

### `func inputIsKnownEscapeEnd(value string) bool`

Logic:
- Recognize common CSI final bytes plus `~` and `u` terminators.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
