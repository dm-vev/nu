# `internal/rpc/jsonl.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 5d9629b
Implementation Comments: Strict LF JSONL reader/writer uses `bufio.Reader` instead of `Scanner`, strips one CR before LF, emits final unterminated records, and wraps marshal/write errors.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Own strict JSONL framing for RPC stdin/stdout.

## Code Style

Use `encoding/json`, `bufio.Reader`, and `io.Writer`. Do not use `Scanner`
because commands can exceed its default token limit. Comments must explain
framing edge cases.

## Functions

### `WriteLine(w io.Writer, value any) error`

Logic:

- Marshal `value` with `encoding/json`.
- Append exactly one `\n`.
- Write once when possible so tests can assert line boundaries.
- Wrap marshal and write errors with operation context.

Acceptance:

- every response/event is one LF-delimited JSON object;
- strings containing Unicode line separators are not split by the writer.

### `ReadLines(r io.Reader, onLine func(string) error) error`

Logic:

- Read until `\n` with `bufio.Reader.ReadString`.
- Strip one trailing `\n`, then one trailing `\r` if present.
- Emit a final unterminated line at EOF.
- Return callback errors immediately.
- Treat plain EOF after all buffered data as success.

Acceptance:

- splits only on LF;
- accepts CRLF input;
- emits the last line without a trailing LF;
- has no scanner token-size ceiling.

Tests:

- `TestRPCJSONLReadLinesStrictLF`
- `TestRPCJSONLReadLinesReturnsCallbackError`
- `TestRPCJSONLWriteLine`
