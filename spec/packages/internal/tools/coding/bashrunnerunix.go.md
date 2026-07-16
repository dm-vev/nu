# `internal/tools/coding/bashrunnerunix.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Unix runner is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Execute shell commands on non-Windows systems with process-group cancellation.

## Code Style

Keep the `!windows` build tag and Unix-only process APIs in this file.

## Owned Logic

- `runCommand` starts `sh -c` in cwd, captures stdout/stderr, and maps run errors to exit codes.
- The shell owns a process group that is killed on context cancellation; deadline expiry reports timeout and exit `-1`.

## Acceptance

- Timed-out child processes do not remain running.
- Streams are returned separately and Unix APIs do not leak into Windows builds.

## Tests

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
