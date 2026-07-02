# `internal/tool/builtin/find.go`

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

Built-in `find` tool.

## Code Style

Use `filepath.WalkDir`. Keep glob semantics documented by tests.

## Functions

### `NewFind(cwd string, opts FindOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `find` and schema.

Acceptance:

- registers name `find` and schema.

### `executeFind(ctx context.Context, args FindArgs, ops FindOps) tool.Result`

Logic:

- Resolve search root under cwd and reject roots outside policy.
- Compile the glob matcher before walking so invalid patterns fail early.
- Walk with `filepath.WalkDir`, skipping ignored directories/files through gitignore matcher.
- Collect relative paths, sort them, and stop when result or byte limits are reached.

Acceptance:

- searches under root path;
- supports glob pattern;
- respects gitignore;
- returns sorted relative paths;
- enforces result and byte limits.

Tests:

- `TestNUF075FindGlob`
- `TestNUF075FindRespectsGitignore`
