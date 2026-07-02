# `internal/tool/bash/runner_unix.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Unix runner owns `sh -c`, cwd execution, stdout/stderr capture, exit-code mapping, and process-group timeout kill.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Run the bash tool command on Unix-like systems without importing Unix-only
`syscall` fields from platform-neutral files.

## Code Style

Use stdlib `exec.CommandContext`. Keep this file Unix-only with a build tag.
No package-level state.

## Functions

### `runCommand(ctx context.Context, cwd string, command string) (stdout string, stderr string, exitCode int, timedOut bool)`

Logic:

- Start `sh -c <command>` in `cwd`.
- Put the shell in its own process group.
- Capture stdout and stderr into buffers.
- On context cancellation, kill the process group.
- Return exit code `0` for success, process exit code for `exec.ExitError`,
  `1` for other run errors, and `-1` for timeout.
- Return timeout state from the context.

Acceptance:

- child processes are killed on timeout;
- stdout and stderr are returned separately;
- no Unix-only imports leak into Windows builds.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
