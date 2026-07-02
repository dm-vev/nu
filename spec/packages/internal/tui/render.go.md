# `internal/tui/render.go`

## Status

Current: IMPLEMENTED
Implementation Commit: pending
Implementation Comments: Renderer builds deterministic frames from state, clamps dimensions, truncates visible text, appends SGR/OSC resets per line, and exposes ANSI stripping for width tests.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render interactive state into deterministic terminal frames without writing to
the terminal.

## Code Style

No terminal globals. Keep ANSI reset and width handling in helper functions.
Comment truncation and resize cache invalidation decisions.

## Types

### `type State struct`

Logic:

- Carry title, cwd, provider, model, status, messages, editor snapshot,
  extension widget lines, and overlay titles.

Acceptance:

- renderer can be tested without an agent.

### `type Frame struct`

Logic:

- Carry rendered lines plus cursor row/column metadata.

Acceptance:

- terminal driver can write exactly this frame.

## Functions

### `Render(state State, width int, height int) Frame`

Logic:

- Clamp width and height to useful minimums.
- Render header, message history, status/widgets, editor, footer, and focused
  overlay title.
- Truncate by visible width after stripping ANSI/OSC sequences.
- Ensure every rendered line ends with SGR reset and OSC 8 reset.
- Keep cursor metadata inside frame bounds.

Acceptance:

- no ANSI-stripped line exceeds terminal width;
- resize changes frame dimensions deterministically.

### `StripANSI(text string) string`

Logic:

- Remove SGR and OSC escape sequences used by the renderer.
- Leave printable text unchanged.

Acceptance:

- tests can measure visible line width.

Tests:

- `TestNUF100RendererDoesNotOverflowWidth`
- `TestNUF100ResizeInvalidatesLayout`
