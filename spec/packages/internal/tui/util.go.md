# `internal/tui/util.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Small helpers for env integers, window title, and non-empty string selection.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Small helpers for env integers, window title, and non-empty string selection.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Functions

### `func envInt(name string, fallback int) int`

Logic:
- Parse a positive integer from an environment variable and fall back when missing or invalid.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func windowTitle(cwd string) string`

Logic:
- Build a terminal title from the current directory basename, falling back to `Nu`.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func firstNonEmpty(values ...string) string`

Logic:
- Return the first trimmed non-empty candidate or an empty string.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
