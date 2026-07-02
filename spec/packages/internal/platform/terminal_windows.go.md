# `internal/platform/terminal_windows.go`

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

Windows terminal mode and process cleanup helpers.

## Code Style

Build-tagged for Windows. Keep behavior parity documented when exact Unix
behavior is impossible.

## Functions

### `EnterRaw(fd uintptr) (RestoreFunc, error)`

Logic:

- Read current console mode for the provided file descriptor.
- Enable virtual terminal input/output and disable line/input echo modes required by TUI.
- Return a restore closure that writes the original mode at most once.
- Wrap Windows syscall errors with the target descriptor and operation.

Acceptance:

- enables required console modes for TUI input;
- restore function is idempotent.

### `KillProcessGroup(cmd *exec.Cmd) error`

Logic:

- Detect whether the command has already exited before sending termination.
- Terminate the job object or process tree abstraction used when spawning bash commands.
- Fall back to killing the child process when no tree handle is available.
- Return nil for already-finished processes and wrap real termination errors.

Acceptance:

- terminates spawned command tree as well as practical on Windows.

Tests:

- covered by Windows CI when available.
