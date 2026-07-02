# `internal/pkgmgr/archive.go`

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

Handle package archive extraction for future binary-distribution packages.

## Code Style

Extraction must prevent path traversal. No lifecycle scripts.

## Functions

### `ExtractArchive(ctx context.Context, r io.Reader, dest string) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Reject absolute paths and `..` traversal.
- Preserve file modes only when safe.
- Create parent directories.

Acceptance:

- rejects absolute paths and `..` traversal;
- preserves file modes only when safe;
- creates parent directories.

Tests:

- `TestPackageArchiveRejectsTraversal`
