# `internal/tool/bash/runner_windows.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Windows runner keeps the bash package compilable on Windows without Unix process-group APIs.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Run the bash tool command on Windows builds without importing Unix-only
`syscall` fields.

## Code Style

Use stdlib `exec.CommandContext`. Keep this file Windows-only with a build tag.
No package-level state.

## Functions

### `runCommand(ctx context.Context, cwd string, command string) (stdout string, stderr string, exitCode int, timedOut bool)`

Logic:

- Start `cmd /C <command>` in `cwd`.
- Capture stdout and stderr into buffers.
- Rely on `CommandContext` cancellation for the spawned process.
- Return exit code `0` for success, process exit code for `exec.ExitError`,
  `1` for other run errors, and `-1` for timeout.
- Return timeout state from the context.

Acceptance:

- package compiles for `GOOS=windows`;
- stdout and stderr are returned separately;
- no Unix-only imports are present.

Tests:

- `GOOS=windows GOARCH=amd64 go test -c ./internal/tool/bash`
