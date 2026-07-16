# `internal/tui/tui_git.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Read git branch metadata for the footer without shelling out.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Read git branch metadata for the footer without shelling out.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.
## Functions

### `func currentGitBranch(cwd string) string`

Logic:
- Read local state directly and return an empty value when unavailable.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
