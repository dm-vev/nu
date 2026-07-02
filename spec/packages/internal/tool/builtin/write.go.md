# `internal/tool/builtin/write.go`

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

Built-in `write` tool.

## Code Style

Write through mutation queue. Create parent directories only when spec says
tool may create files.

## Functions

### `NewWrite(cwd string, opts WriteOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `write`, schema, and sequential mutation behavior.

Acceptance:

- registers name `write`, schema, and sequential mutation behavior.

### `executeWrite(ctx context.Context, args WriteArgs, ops WriteOps) tool.Result`

Logic:

- Resolve the target path relative to the tool cwd and reject unsafe traversal if policy forbids it.
- Validate content is UTF-8 when text mode is requested.
- Acquire the per-file mutation lock before creating parents or writing.
- Write the file through injected ops and return canonical path plus byte count in details.

Acceptance:

- writes UTF-8 content;
- creates or overwrites target file;
- returns path and byte count details.

Tests:

- `TestNUF071WriteCreatesFile`
