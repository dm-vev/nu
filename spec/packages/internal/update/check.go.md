# `internal/update/check.go`

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

Check available Nu and package updates.

## Code Style

Network calls use injected HTTP client. Offline mode short-circuits before any
request.

## Functions

### `Check(ctx context.Context, opts CheckOptions) (Report, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Checks self and packages when allowed.
- Skips all network when offline.
- Return structured report.

Acceptance:

- checks self and packages when allowed;
- skips all network when offline;
- returns structured report.

Tests:

- `TestNUF182OfflineSkipsNetworkChecks`
