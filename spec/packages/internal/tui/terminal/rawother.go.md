# `internal/tui/terminal/rawother.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Provide non-Unix raw-mode fallback.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Provide non-Unix raw-mode fallback.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.
## Functions

### `func (t *Terminal) EnableRaw() (func() error, bool, error)`

Logic:
- EnableRaw is a no-op on unsupported platforms.

Acceptance:
- Terminal state is restored or cleanup is returned on every successful setup path.
