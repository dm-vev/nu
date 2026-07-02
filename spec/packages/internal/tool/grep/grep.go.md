# `internal/tool/grep/grep.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Searches files under cwd with literal or regexp matching.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `grep` built-in tool.

## Code Style

Stdlib walk, scanner, and regexp only. Phase 2 `.gitignore` support is simple
root-pattern handling from `toolkit`.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode `pattern`, `literal`, `ignore_case`, `glob`, `root`, and `limit`.
- Reject missing pattern.
- Default limit to 100.
- Resolve root under cwd.
- Compile regexp matcher or create literal matcher.
- Walk root, skipping `.git` and root `.gitignore` patterns.
- For matching files, scan lines and emit `path:line:text`.
- Stop after limit and return sorted matches as JSON.

Acceptance:

- supports literal and regexp matching;
- supports ignore-case matching;
- respects root `.gitignore`;
- rejects invalid regexp.

Tests:

- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`
- `TestGrepIgnoreCase`
- `TestGrepRejectsInvalidRegex`
