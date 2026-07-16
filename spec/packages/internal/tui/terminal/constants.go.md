# `internal/tui/terminal/constants.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define terminal control sequences.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define terminal control sequences.

## Code Style

Use stdlib syscalls and injected IO. Restore terminal state on every successful raw enable. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- No test depends on the developer terminal unless guarded by injected IO or fallbacks.

## Types And Constants

### `const (...)`

Logic:
- Define synchronized output, cursor visibility, bracketed paste, and mouse-reporting cleanup control sequences emitted by the terminal/engine packages.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

This file declares data/constants only; behavior is exercised through files in the same package.
