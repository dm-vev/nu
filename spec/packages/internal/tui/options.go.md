# `internal/tui/options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Normalize externally supplied interactive app options.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Normalize externally supplied interactive app options.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.

## Types And Constants

### `type AppOptions struct {`

Logic:
- AppOptions configures interactive mode.
- Carry current provider/model labels, session id/name labels, visible model candidates, and an optional model selection callback.
- Carry `ASCII` as an explicit limited-character override for terminals that cannot render Nu's Unicode TUI glyphs.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

### `const defaultContext = 128000`

Logic:
- Provide the fallback model context window used when options do not supply one.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func normalizeOptions(opts AppOptions) AppOptions`

Logic:
- Clamp invalid options to deterministic defaults and fill nil callbacks with safe behavior.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func limitedCharset(opts AppOptions) bool`

Logic:
- Return true when `AppOptions.ASCII` is set, `NU_TUI_ASCII` is truthy, or `TERM` names a limited terminal such as `linux`, `dumb`, `vt100`, `vt102`, or `ansi`.
- Keep full Unicode rendering for modern terminals by default.

Acceptance:
- `TestLimitedCharsetDetectsOptionEnvAndTerm` fails if explicit or detected ASCII mode stops working.

### `func statusFrames(opts AppOptions) []string`

Logic:
- Return ASCII spinner frames `-`, `\`, `|`, `/` for limited terminals.
- Return nil for normal terminals so the status component uses its default Claude-like frames.

Acceptance:
- `TestTUIAppUsesLimitedCharsetWhenRequested` fails if limited terminals receive Unicode status frames.
