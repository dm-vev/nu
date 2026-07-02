# `internal/tui/theme.go`

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

Apply theme colors/styles to TUI rendering.

## Code Style

Theme lookup returns default-safe styles. Rendering code must not panic on
missing optional theme fields.

## Functions

### `ResolveTheme(themes []Theme, name string, terminal TerminalTheme) Theme`

Logic:

- If an explicit theme name is provided, match it by stable theme id first, then display name.
- For `auto`, select dark/light from terminal theme detection when available.
- Fall back to the built-in dark theme when no requested theme is found.
- Return a complete theme with every token filled from fallback defaults.

Acceptance:

- resolves explicit name, auto dark/light, and fallback theme.

### `Style(theme Theme, token Token, text string) string`

Logic:

- Look up the token style, falling back to plain text for unknown tokens.
- Wrap text with ANSI start sequence and reset sequence exactly once.
- Avoid nesting resets inside the text; caller owns already-styled fragments.
- Return input text unchanged when styling is disabled.

Acceptance:

- applies ANSI style and reset safely.

Tests:

- `TestTUIResolveThemeFallback`
