# `internal/tui/thinking_options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines thinking padding and styles.

## TODO

- [x] File exists in the split component architecture.
- [x] Nil callbacks normalize safely.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Configure thinking rendering without coupling it to app theme functions.

## Functions

### `func thinkingNormalizeOptions(opts ThinkingOptions) ThinkingOptions`

Logic:
- Clamp padding and default missing style callbacks.

Acceptance:
- Rendering works when callers only set `TextStyle`.

### `func thinkingIdentity(value string) string`

Logic:
- Return text unchanged.

Acceptance:
- Used as the default style callback.
