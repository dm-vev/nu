# `internal/tool/ls/ls.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Lists one directory under cwd.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `ls` built-in tool.

## Code Style

Use `os.ReadDir`. Include dotfiles. Keep output stable and sorted.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode optional `path`, defaulting to `.`.
- Resolve directory under cwd and reject escapes.
- Check cancellation before read.
- Read directory entries.
- Append `/` to directory names.
- Sort entries.
- Return JSON `entries` list with truncation.

Acceptance:

- lists sorted entries with directories marked;
- includes dotfiles;
- rejects non-directories and cwd escapes.

Tests:

- `TestNUF076LsSortedWithDirs`
- `TestNUF076LsRejectsNonDirectory`
- `TestLsRejectsPathEscape`
