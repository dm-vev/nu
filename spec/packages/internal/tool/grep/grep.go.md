# `internal/tool/grep/grep.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Grep tool lives in its own subpackage with literal, regexp, ignore-case, gitignore, and long-line truncation tests.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

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
- For matching files, scan lines with an explicit large token limit.
- Emit `path:line:text`, truncating very long matching line text before it
  enters the tool result.
- Stop after limit and return sorted matches as JSON.

Acceptance:

- supports literal and regexp matching;
- supports ignore-case matching;
- respects root `.gitignore`;
- handles long matching lines without scanner token errors;
- truncates very long matching line text;
- rejects invalid regexp.

Tests:

- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`
- `TestGrepIgnoreCase`
- `TestGrepTruncatesLongMatchingLine`
- `TestGrepRejectsInvalidRegex`
