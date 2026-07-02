# `internal/pkgmgr/source.go`

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

Parse and normalize package source strings.

## Code Style

Pure parser. No network or filesystem.

## Functions

### `ParseSource(input string) (Source, error)`

Logic:

- Trim surrounding whitespace and reject an empty source string.
- Classify source as local path, `npm:` alias, `git:` shorthand, or protocol git URL.
- Split pinned refs from the source without losing URL credentials or path components.
- Compute normalized package identity from source kind, repo/path, package name, and ref.

Acceptance:

- supports local paths, `npm:`, `git:`, and protocol git URLs;
- extracts pinned refs when present;
- normalizes package identity for deduplication.

Tests:

- `TestPackageParseLocalSource`
- `TestPackageParseGitPinnedRef`
