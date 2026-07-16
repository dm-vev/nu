# `internal/tools/coding/bash_runner_windows.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Windows runner is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Execute shell commands on Windows without Unix-only process APIs.

## Code Style

Keep the `windows` build tag and use stdlib `exec.CommandContext`.

## Owned Logic

- `runCommand` starts `cmd /C` in cwd and captures stdout/stderr.
- It maps success, process errors, other errors, and deadline expiry to stable exit/timed-out values.

## Acceptance

- `internal/tools/coding` compiles for Windows.
- Streams remain separate and timeout reports exit `-1`.

## Tests

- `GOOS=windows GOARCH=amd64 go test -c ./internal/tools/coding`
