# `internal/tui/renderer.go`

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

Render component tree to terminal frames.
Implements frame rules from `spec/protocols/tui-rendering.md`.

## Code Style

Renderer owns width enforcement. Components should behave, renderer still
guards.

## Functions

### `Render(root Component, size Size) Frame`

Logic:

- Validate terminal width and height are positive.
- Ask root component to render at requested width.
- Strip ANSI for width accounting while preserving original styled text.
- Hard-wrap or truncate any component line that exceeds width after ANSI-strip.
- Append SGR reset and OSC 8 reset to each physical line.
- Track cursor marker metadata before removing marker from visible output.
- On resize generation change, invalidate root before rendering.
- Return frame lines and cursor metadata without writing to terminal.

Acceptance:

- no output line exceeds terminal width;
- appends reset escapes per line;
- handles resize invalidation.

### `Diff(prev, next Frame) []WriteOp`

Logic:

- Compare previous and next frame line by line.
- Emit cursor moves and line writes for changed lines.
- Clear stale tail content when a new line is shorter than previous line.
- Clear removed rows when next frame has fewer lines and clear-on-shrink is
  active.
- End with cursor placement/hide/show operation from next frame metadata.

Acceptance:

- emits minimal-enough writes without sacrificing correctness.

Tests:

- `TestNUF100RendererDoesNotOverflowWidth`
- `TestNUF100ResizeInvalidatesLayout`
