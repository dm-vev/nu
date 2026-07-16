# `internal/tui/terminal_raw_unix.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Enable/restore raw mode on Unix TTYs.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Enable/restore raw mode on Unix TTYs.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.

## Types And Constants

### `type terminalFdReader interface {`

Logic:
- Narrowly detect injected stdin values that expose a file descriptor.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func (t *Terminal) EnableRaw() (func() error, bool, error)`

Logic:
- EnableRaw enables raw mode when stdin is a TTY.
- It reads the current termios state, disables canonical/echo/signal translation flags, writes the raw state, and returns a restore closure that writes the saved state back.

Acceptance:
- Terminal state is restored or cleanup is returned on every successful setup path.
- Non-TTY stdin returns `(nil, false, nil)` so tests and pipes use line mode.

### `func terminalIoctlTermios(fd uintptr, request uintptr, state *syscall.Termios) error`

Logic:
- Call the shared ioctl wrapper for termios-specific state reads/writes.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func terminalIoctl(fd uintptr, request uintptr, pointer unsafe.Pointer) error`

Logic:
- Execute `SYS_IOCTL` and return `errno` as an error when the syscall fails.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
