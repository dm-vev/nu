# `internal/tool/truncation.go`

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

Shared line/byte truncation helpers.

## Code Style

Pure functions. Count bytes as UTF-8 bytes and never split invalid UTF-8 into
output.

## Functions

### `TruncateTail(text string, opts Options) Result`

Logic:

- Count bytes and display lines before truncation so details report original totals.
- Keep the end of the text when max bytes or max lines is exceeded.
- Cut only on valid UTF-8 boundaries and preserve line order.
- Return reason, omitted byte count, omitted line count, and whether truncation occurred.

Acceptance:

- respects max lines and max bytes;
- reports total lines/bytes and truncation reason.

### `TruncateHead(text string, opts Options) Result`

Logic:

- Count bytes and display lines before truncation so details report original totals.
- Keep the beginning of the text when max bytes or max lines is exceeded.
- Cut only on valid UTF-8 boundaries and preserve line order.
- Return reason, omitted byte count, omitted line count, and whether truncation occurred.

Acceptance:

- keeps tail/head semantics distinct and tested.

Tests:

- `TestTruncateTailByLines`
- `TestTruncateTailByBytes`
