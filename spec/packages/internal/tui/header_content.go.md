# `internal/tui/header_content.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Build compact and expanded header text.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Build compact and expanded header text.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Functions

### `func (h *Header) content() string`

Logic:
- Compose logo, help text, startup hint, onboarding text, and the minimal loaded context block based on expansion state.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (h *Header) compactHelp() string`

Logic:
- Return one-line Pi-style compact key help with muted separators.
- Use the configured separator so limited terminals can render ASCII-only help.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (h *Header) expandedHelp() string`

Logic:
- Return multi-line help covering interrupt, clear, exit, commands, bash, and expansion actions.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (h *Header) onboarding() string`

Logic:
- Return the dim onboarding sentence shown under startup help.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (h *Header) contextBlock() string`

Logic:
- Return the minimal loaded resources block currently shown at startup: `[Context]` and `AGENTS.md`.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
