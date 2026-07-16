# `internal/tui/header_options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Normalize externally supplied interactive app options.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Normalize externally supplied interactive app options.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type HeaderOptions struct {`

Logic:
- Options configures the startup header.
- Carry `HelpSeparator` so limited terminals can replace Unicode compact-help separators.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func headerNormalizeOptions(opts HeaderOptions) HeaderOptions`

Logic:
- Clamp invalid options to deterministic defaults and fill nil callbacks with safe behavior.
- Default the compact help separator to ` · ` when no override is supplied.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func headerIdentity(value string) string`

Logic:
- Return the provided value unchanged as the default header style callback.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
