
# `internal/update/update.go`

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

Apply self and package updates.

## Code Style

All external commands are injected. Never overwrite the running binary without a
verified replacement path.

## Functions

### `Apply(ctx context.Context, report Report, opts ApplyOptions) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Updates selected targets only.
- Write progress events.
- Leaves current install untouched on failed replacement.

Acceptance:

- updates selected targets only;
- writes progress events;
- leaves current install untouched on failed replacement.

Tests:

- `TestUpdateApplySelectedTargetsOnly`
- `TestNUF182UpdateLeavesInstallUntouchedOnFailure`
