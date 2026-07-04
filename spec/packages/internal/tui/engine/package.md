# `internal/tui/engine`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Pi-style TUI engine: render a component tree, extract cursor markers, apply
line resets, and write synchronized full/differential terminal updates.

## Files

- `engine.go`: engine state and constructor.
- `options.go`: renderer options.
- `render.go`: render decision path.
- `diff.go`: full and differential writes.
- `cursor.go`: cursor extraction and positioning.
- `lifecycle.go`: start/stop terminal state.
- `engine_test.go`: synchronized render/diff smoke test.

## Acceptance Criteria

- One render call writes one synchronized output buffer.
- Ordinary streaming updates use diff repaint, not clear-screen replay.
- Full render is reserved for first paint and size changes.
- First paint clears the screen.
- Flexible fillers anchor editor/footer to the terminal bottom.
- Stop leaves the terminal below the rendered frame and shows the cursor.
