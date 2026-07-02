# `internal/tool/find/find.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 6ec7970
Implementation Comments: Find tool lives in its own subpackage with glob, limit, gitignore, and cwd escape tests.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `find` built-in tool.

## Code Style

Stdlib `filepath.WalkDir` only. Return sorted slash-form relative paths.

## Functions

### `Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode optional `root`, `glob`, and `limit`.
- Default root to `.` and limit to 100.
- Resolve root under cwd.
- Walk files only.
- Skip `.git` and root `.gitignore` patterns.
- Match glob against relative path or base name.
- Stop at limit and return sorted JSON paths.

Acceptance:

- finds files by glob;
- respects root `.gitignore`;
- enforces limit;
- rejects cwd escapes.

Tests:

- `TestNUF075FindGlob`
- `TestNUF075FindRespectsGitignore`
- `TestFindEnforcesLimit`
- `TestFindRejectsPathEscape`
