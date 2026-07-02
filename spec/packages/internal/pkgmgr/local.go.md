# `internal/pkgmgr/local.go`

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

Resolve local package resources.

## Code Style

Do not copy local packages. Resolve paths relative to settings file.

## Functions

### `ResolveLocal(ctx context.Context, src Source, base string) (ResolvedPackage, error)`

Logic:

- Resolve the local source path relative to `base` and clean it.
- Stat the path through injected filesystem operations.
- Treat a file as one extension entry and a directory as a package root with manifest discovery.
- Return a path-qualified missing-source error when the path does not exist.

Acceptance:

- accepts file as single extension;
- accepts directory as package root;
- rejects missing paths with clear error.

Tests:

- `TestNUF150LocalPackageDiscovery`
