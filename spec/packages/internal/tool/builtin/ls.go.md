# `internal/tool/builtin/ls.go`

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

Built-in `ls` tool.

## Code Style

No shelling out. Use filesystem APIs.

## Functions

### `NewLS(cwd string, opts LSOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `ls` and schema.

Acceptance:

- registers name `ls` and schema.

### `executeLS(ctx context.Context, args LSArgs, ops LSOps) tool.Result`

Logic:

- Resolve the directory path relative to cwd and check context before reading.
- Stat the path and return a tool error when it is not a directory.
- Read entries including dotfiles, sort by display name, and mark directories with `/`.
- Apply entry and byte limits while reporting truncation details.

Acceptance:

- rejects non-directories;
- includes dotfiles;
- sorts alphabetically;
- appends `/` for directories;
- enforces entry and byte limits.

Tests:

- `TestNUF076LsSortedWithDirs`
- `TestNUF076LsRejectsNonDirectory`
