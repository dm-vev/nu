# `internal/tui/footer_options.go`

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

### `type FooterOptions struct {`

Logic:
- Options configures the two-line Pi-style footer.
- It carries cwd/home/branch identity, provider/model labels, right-side notice, used context estimate, context window, and style callbacks.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func footerNormalizeOptions(opts FooterOptions) FooterOptions`

Logic:
- Clamp invalid options to deterministic defaults and fill nil callbacks with safe behavior.
- Clamp negative used context to zero and default missing context windows to 128k.
- Default notice styling to the dim footer style when no explicit style is supplied.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func footerIdentity(value string) string`

Logic:
- Return the provided value unchanged as the default footer style callback.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
