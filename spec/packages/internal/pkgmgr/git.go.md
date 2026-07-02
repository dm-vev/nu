# `internal/pkgmgr/git.go`

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

Install and reconcile git package sources.

## Code Style

All git commands go through injected runner. Never run destructive git commands
outside package clone directory.

## Functions

### `InstallGit(ctx context.Context, src Source, dest string, runner Runner) (ResolvedPackage, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Clone missing repo.
- Fetch and checks out pinned ref.
- Disable credential prompts in non-interactive mode when configured.

Acceptance:

- clones missing repo;
- fetches and checks out pinned ref;
- disables credential prompts in non-interactive mode when configured.

### `ReconcileGit(ctx context.Context, pkg ResolvedPackage, runner Runner) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Reset only package clone dir to configured ref.
- Clean only package clone dir.

Acceptance:

- resets only package clone dir to configured ref;
- cleans only package clone dir.

Tests:

- `TestNUF150GitPackagePinnedRef`
