# `internal/session/resume.go`

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

Find, continue, fork, and clone sessions.

## Code Style

All filesystem scanning is bounded and cancellable. Sorting is deterministic.

## Functions

### `FindLatest(ctx context.Context, store *Store, cwd string) (Ref, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Return most recent session for cwd.
- Ignores malformed session files with diagnostics when possible.

Acceptance:

- returns most recent session for cwd;
- ignores malformed session files with diagnostics when possible.

### `ResolveRef(ctx context.Context, store *Store, input string) (Ref, error)`

Logic:

- Check whether input names an existing session file path first.
- Resolve exact session id matches before partial id matching.
- For partial ids, scan candidate session refs deterministically and collect all matches.
- Return an ambiguous-ref error listing candidates when more than one ref matches.

Acceptance:

- supports path, full id, and partial id;
- errors on ambiguous partial ids.

### `Fork(ctx context.Context, store *Store, source Ref, entryID string) (Ref, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Start a new file from selected user entry context.

Acceptance:

- starts a new file from selected user entry context.

### `Clone(ctx context.Context, store *Store, source Ref) (Ref, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Copies active branch into a new session file.

Acceptance:

- copies active branch into a new session file.

Tests:

- `TestNUF081ContinueLatestByCWD`
- `TestNUF081ForkStartsNewFileFromUserEntry`
- `TestNUF081CloneCopiesActiveBranch`
