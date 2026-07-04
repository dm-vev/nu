# `internal/tui/components/usermessage/options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Normalizes user Markdown and background style callbacks.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Nil callbacks normalize safely.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Configure user message Markdown and background rendering.

## Types

### `type Options struct`

Logic:
- Carry padding, text/strong/emphasis/code styles, and optional background style.

Acceptance:
- User message palette stays injected from `internal/tui`.

## Functions

### `func normalizeOptions(opts Options) Options`

Logic:
- Clamp padding and fill missing text styles from `TextStyle`.

Acceptance:
- User message rendering works with partial options.

### `func identity(value string) string`

Logic:
- Return text unchanged.

Acceptance:
- Used as the default style callback.
