# `internal/platform/terminal_unix.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Unix terminal raw mode, process groups, and signal helpers.

## Code Style

Build-tagged for Unix. Keep syscalls isolated and tested behind interfaces where
possible.

## Functions

### `EnterRaw(fd uintptr) (RestoreFunc, error)`

Logic:

- Read current termios for the provided file descriptor.
- Disable canonical input, echo, and signal processing flags required by the TUI.
- Apply raw mode with `TCSANOW` and return an idempotent restore closure.
- Wrap syscall errors with operation and descriptor context.

Acceptance:

- switches terminal to raw mode;
- restore function is idempotent.

### `KillProcessGroup(cmd *exec.Cmd) error`

Logic:

- Look up the process group id for the spawned command.
- Send the configured termination signal to the negative pgid so children receive it too.
- Treat already-exited or missing processes as a successful cleanup.
- Return path/pgid-qualified errors for real signal failures.

Acceptance:

- terminates command process group for bash timeout/abort.

Tests:

- covered by platform integration tests where available.
