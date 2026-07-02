# `internal/tool/find/find.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Finds files under cwd using glob and root `.gitignore` handling.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

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
