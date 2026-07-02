# `internal/tool/builtin/mutation.go`

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

Serialize file mutations for write/edit tools.

## Code Style

Small per-path lock map. Remove locks when idle to avoid unbounded growth.

## Functions

### `WithFileMutation(ctx context.Context, path string, fn func() error) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Serializes mutations on the same canonical path.
- Allow different files to mutate concurrently.
- Releases lock on panic/error/cancel.

Acceptance:

- serializes mutations on the same canonical path;
- allows different files to mutate concurrently;
- releases lock on panic/error/cancel.

Tests:

- `TestNUF071ConcurrentWritesSamePathSerialize`
