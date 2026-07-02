# `internal/tool/bash/bash.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Runs shell commands with timeout, captured output, and truncation persistence.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `bash` built-in tool.

## Code Style

Use `exec.CommandContext` and `sh -c`. On Unix, kill the process group on
timeout so child processes do not outlive the shell.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode `command` and optional `timeout_ms`.
- Reject empty commands.
- Create timeout context when requested.
- Start `sh -c` in cwd.
- Capture stdout and stderr.
- On timeout, kill process group and mark `timed_out`.
- Return JSON with stdout, stderr, exit code, timed_out, output, truncated, and
  optional `full_output_path`.
- Persist full output to temp file when displayed output is truncated.

Acceptance:

- captures stdout, stderr, and exit code;
- kills timed-out commands promptly;
- persists full output when truncating display output.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
- `TestBashRejectsEmptyCommand`
