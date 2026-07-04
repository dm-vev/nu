# `internal/tool/bash/bash.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Bash tool lives in its own subpackage; platform runners keep Unix process-group timeout handling and Windows compilation.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `bash` built-in tool.

## Code Style

Keep argument decoding, timeout setup, truncation, and JSON result formatting in
`Run`. Delegate only process execution to platform files.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode `command` and optional `timeout_ms`.
- Reject empty commands.
- Reject interactive `sudo` before starting a process; allow only explicit
  non-interactive forms `sudo -n`, `sudo -S`, or `sudo --non-interactive`.
- Create timeout context when requested.
- Call the platform runner in cwd.
- Capture stdout, stderr, exit code, and timeout state from the runner.
- Return JSON with stdout, stderr, exit code, timed_out, output, truncated, and
  optional `full_output_path`.
- Persist full output to temp file when displayed output is truncated.

Acceptance:

- captures stdout, stderr, and exit code;
- kills timed-out commands promptly;
- persists full output when truncating display output.
- never lets a password prompt take over the TUI input row.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
- `TestBashRejectsEmptyCommand`
- `TestBashRejectsInteractiveSudo`
- `TestBashAllowsNonInteractiveSudoForms`

### `usesInteractiveSudo(command string) bool`

Logic:

- Scan shell words for a `sudo` command token.
- Return true unless the sudo invocation carries `-n`, `-S`, or
  `--non-interactive` before the sudo subcommand.

Acceptance:

- Plain `sudo true` is blocked before process start.
- Non-interactive sudo forms are passed through to the platform runner.

### `sudoNonInteractiveFlag(value string) bool`

Logic:

- Recognize short sudo option groups containing `n` or `S`, and the long
  `--non-interactive` option.

Acceptance:

- `--preserve-env` and other unrelated long flags do not bypass the guard.

Compile checks:

- `GOOS=windows GOARCH=amd64 go test -c ./internal/tool/bash`
