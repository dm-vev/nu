# `internal/tui/terminal/size_unix.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Query terminal size on Unix writers with file descriptors.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Query terminal size on Unix writers with file descriptors.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.

## Functions

### `func querySize(target any, fallbackWidth int, fallbackHeight int) (int, int)`

Logic:
- Use `TIOCGWINSZ` when the writer exposes a file descriptor; otherwise return fallback dimensions.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
