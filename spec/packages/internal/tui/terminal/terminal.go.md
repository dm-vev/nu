# `internal/tui/terminal/terminal.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Wrap injected stdin/stdout and terminal dimensions.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Wrap injected stdin/stdout and terminal dimensions.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.

## Types And Constants

### `type Terminal struct {`

Logic:
- Terminal owns raw terminal IO and dimensions.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New(stdin io.Reader, stdout io.Writer, width int, height int) *Terminal`

Logic:
- New creates a terminal wrapper.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) Stdin() io.Reader`

Logic:
- Stdin returns the input reader.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (t *Terminal) Size() (int, int)`

Logic:
- Size returns current terminal dimensions.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
