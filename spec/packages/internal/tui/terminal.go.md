# `internal/tui/terminal.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Terminal writer owns Pi-like TTY setup, synchronized output, first-frame write, home repaint for later frames, fixed-width stale-line clearing, cursor positioning back to the editor row, close-below-frame restore bytes, and append fallback for pipe tests.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Own terminal frame writes so interactive rendering repaints one screen instead
of appending every frame to scrollback.

## Code Style

Use only ANSI escape sequences and injected writers. No new terminal dependency.
Keep fallback mode append-only for pipe tests. Comment the repaint sequence.

## Types

### `type Terminal struct`

Logic:

- Store writer, repaint flag, first-frame state, and previous frame line count.
- In repaint mode, use bracketed paste, cursor hide, terminal title, synchronized
  output, and cursor-home repaint for subsequent frames.
- Clear stale rows by writing fixed-width blank lines when the new frame is
  shorter than the previous frame.
- In append mode, write lines plainly for deterministic tests.

Acceptance:

- real interactive output updates in place;
- tests can still capture frame text without terminal control noise unless they
  opt into repaint mode.

## Functions

### `NewTerminal(w io.Writer, repaint bool) *Terminal`

Logic:

- Normalize nil writer to `io.Discard`.
- Store repaint behavior.

Acceptance:

- no process globals.

### `(*Terminal) Draw(frame Frame) error`

Logic:

- In repaint mode, write setup bytes before the first frame.
- For later frames, write synchronized-output start plus `CSI H`, then frame
  lines separated by CRLF without `CSI 2J`.
- Move the cursor back to `Frame.CursorRow`/`Frame.CursorCol` after writing the
  frame.
- Write each append-mode frame line followed by newline.
- Keep writes ordered and wrap writer errors.

Acceptance:

- repeated draws do not append duplicate visible frames in a terminal;
- repeated draws do not clear the whole screen with `CSI 2J`;
- returned errors include operation context.

### `(*Terminal) Close() error`

Logic:

- Move the cursor below the last rendered frame, then restore synchronized
  output, cursor visibility, and bracketed paste state when repaint mode was
  started.
- Do nothing in append mode.

Acceptance:

- clean interactive exits and provider errors leave following stderr/stdout
  below the TUI frame.

Tests:

- `TestTerminalDrawRepaintsWithANSI`
- `TestTerminalDrawAppendModePlain`
- `TestTerminalCloseMovesBelowFrame`
