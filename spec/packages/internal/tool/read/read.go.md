# `internal/tool/read/read.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Reads text and image files under cwd.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `read` built-in tool.

## Code Style

No global state. Decode JSON args locally, use `toolkit.ResolveUnder`, and
return JSON tool results.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode `path`, optional `offset`, and optional `limit`.
- Resolve path under cwd.
- Check cancellation before filesystem reads.
- Reject directories.
- Read file bytes.
- If extension is supported image type, return `path`, `mime_type`, and base64
  `data`.
- For text, validate offset, apply offset and limit, truncate to `maxBytes`,
  and return `path`, `content`, and `truncated`.

Acceptance:

- reads text with offset/limit;
- truncates large text;
- returns image attachments;
- rejects cwd escapes and directories.

Tests:

- `TestNUF070ReadTextWithOffsetLimit`
- `TestNUF070ReadTruncatesLargeFile`
- `TestNUF070ReadImageAttachment`
- `TestReadRejectsPathEscape`
