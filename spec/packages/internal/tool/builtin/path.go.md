# `internal/tool/builtin/path.go`

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

Resolve user paths for built-in filesystem tools.

## Code Style

Use `filepath`. Keep Pi/macOS compatibility fallbacks localized here.

## Functions

### `ResolveToCWD(cwd, input string) (string, error)`

Logic:

- Reject empty input with a field-qualified error.
- Strip a leading `@` only as a CLI file-reference marker; keep the remaining
  bytes unchanged before path expansion.
- Expand `~` against the caller-provided home/cwd context, not `os.UserHomeDir`.
- Resolve relative paths under `cwd`, clean the result with `filepath.Clean`,
  and return the path even if it does not exist.

Acceptance:

- handles absolute, relative, `~`, and `@`-prefixed file references;
- cleans paths without forcing existence.

### `ResolveReadPath(cwd, input string, fs FS) (string, error)`

Logic:

- Resolve the primary path through `ResolveToCWD` and check it with the injected
  filesystem.
- If the primary path is missing, generate the Pi/macOS screenshot variants:
  non-breaking spaces, NFD normalization, and curly-quote substitutions.
- Return the first variant that exists; if none exist, return the original
  resolved path with a not-found error that lists attempted variants.

Acceptance:

- tries macOS screenshot NBSP, NFD, and curly quote variants when exact path is
  missing.

Tests:

- `TestBuiltinResolveToCWD`
- `TestBuiltinResolveReadPathVariants`
