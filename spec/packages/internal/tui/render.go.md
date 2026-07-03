# `internal/tui/render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Renderer builds Pi-style deterministic frames: startup header, compact help, context block, chat/status/widgets, bordered editor, cwd footer, context/model footer, dark green/gray/white palette, visible-width truncation, fixed-width padding, SGR/OSC reset suffixes, and cursor metadata.

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

- Carry title, version, cwd, home, git branch, provider, model display label,
  context window, auto-compaction flag, context file list, status, messages,
  editor snapshot, extension widget lines, and overlay titles.

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
- Render the Pi-compatible startup surface: blank top spacer, `Nu v...` logo,
  compact keybinding help, startup hint, onboarding line, `[Context]` block,
  message history, status/widgets, horizontal editor borders, blank/typed editor
  row, cwd/branch footer, and right-aligned provider/model footer.
- Apply the built-in dark green, black-terminal, gray, and white/text palette;
  avoid purple/cyan as primary UI colors.
- Truncate by visible width after stripping ANSI/OSC sequences.
- Pad each rendered line to terminal width before the SGR reset and OSC 8 reset
  so repaint mode clears previous longer content without `CSI 2J`.
- Keep cursor metadata inside frame bounds.

Acceptance:

- no ANSI-stripped line exceeds terminal width;
- resize changes frame dimensions deterministically.
- rendered frames use only the built-in palette colors.

### `StripANSI(text string) string`

Logic:

- Remove SGR and OSC escape sequences used by the renderer.
- Leave printable text unchanged.

Acceptance:

- tests can measure visible line width.

Tests:

- `TestNUF100RendererDoesNotOverflowWidth`
- `TestNUF100ResizeInvalidatesLayout`
- `TestNUF100RendererUsesDarkGreenPalette`
- `TestNUF100RendererTruncatesVisibleTextWithoutBreakingANSI`
