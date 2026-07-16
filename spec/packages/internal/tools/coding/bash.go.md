# `internal/tools/coding/bash.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Bash behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Validate and execute one shell command with timeout, output limits, and sudo safeguards.

## Code Style

Keep platform execution in runner files and return one JSON `Result`.

## Owned Logic

- `RunBash` decodes command/timeout, rejects empty or interactive sudo, runs in cwd, captures streams/status, truncates display output, and persists full truncated output.
- Sudo helpers allow only explicit `-n`, `-S`, or `--non-interactive` forms.

## Acceptance

- Timeout, stdout, stderr, exit code, and truncation metadata are accurate.
- Interactive password prompts never take over TUI input.
- Full output remains available when display output is truncated.

## Tests

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
- `TestBashRejectsInteractiveSudo`
