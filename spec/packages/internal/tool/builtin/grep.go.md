# `internal/tool/builtin/grep.go`

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

Built-in `grep` tool.

## Code Style

Prefer stdlib regex and filesystem walking. Keep gitignore matching behind a
small matcher interface.

## Functions

### `NewGrep(cwd string, opts GrepOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `grep` and schema.

Acceptance:

- registers name `grep` and schema.

### `executeGrep(ctx context.Context, args GrepArgs, ops GrepOps) tool.Result`

Logic:

- Resolve root path and compile regex or literal matcher before walking.
- Build file include/exclude filters and gitignore matcher from args/options.
- Scan files line by line with context lines while respecting cancellation.
- Format matches as `file:line:content`, truncate long lines, and stop at configured limits.

Acceptance:

- supports regex/literal, ignore-case, glob, context lines, limit, gitignore;
- returns file:line formatted results;
- truncates long lines.

Tests:

- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`
