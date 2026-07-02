# `internal/testkit/fs.go`

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

Filesystem and environment helpers for tests.

## Code Style

Use temp dirs. Never touch real home.

## Functions

### `NewHome(t testing.TB) Home`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Create temp home and cwd.
- Return env map with no provider keys unless explicitly set.

Acceptance:

- creates temp home and cwd;
- returns env map with no provider keys unless explicitly set.

### `WriteFile(t testing.TB, path, content string)`

Logic:

- Call `t.Helper()` before any filesystem operation.
- Create all parent directories under the test home or explicit temp root.
- Write UTF-8 fixture content with deterministic permissions.
- Fail the test immediately with the target path when mkdir or write fails.

Acceptance:

- writes fixture files with parent dirs.

Tests:

- helper is used by config/session/resource tests.
