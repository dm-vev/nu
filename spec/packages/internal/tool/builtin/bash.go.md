# `internal/tool/builtin/bash.go`

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

Built-in `bash` tool.

## Code Style

Use `exec.CommandContext`. Process group cleanup is platform-specific and
delegated to `internal/platform`.

## Functions

### `NewBash(cwd string, opts BashOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `bash`, schema, and update streaming.

Acceptance:

- registers name `bash`, schema, and update streaming.

### `executeBash(ctx context.Context, args BashArgs, runner Runner, acc OutputAccumulator) tool.Result`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Reject empty command before spawning a process.
- Derive execution context from caller context plus optional timeout.
- Build shell argv from configured shell path and shell command prefix.
- Start process in cwd with configured environment and process-group settings
  from `internal/platform`.
- Stream stdout and stderr into one ordered output accumulator, tagging source
  when details need it.
- Send periodic tool execution updates from accumulator snapshots.
- On timeout or cancellation, terminate process group and mark result cancelled.
- Wait for process exit and capture exit code when available.
- Snapshot accumulator with `persistIfTruncated=true`.
- Return text content, exit code, cancellation flag, truncation details, and full
  output temp path when present.

Acceptance:

- runs command in cwd;
- captures stdout and stderr;
- honors timeout;
- returns exit code and cancellation state;
- truncates display output and stores full output when truncated.

Tests:

- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
