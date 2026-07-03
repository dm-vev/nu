# TUI Rendering And Input Contract

## Purpose

Keep terminal behavior deterministic and testable while matching Pi's terminal
surface closely enough for byte-level startup checks.

## Frame Rules

- Renderer input is a component tree plus terminal size.
- Renderer output is a frame: ordered display lines plus cursor metadata.
- No rendered line may exceed terminal width after ANSI escape stripping.
- Each rendered line ends with SGR reset and OSC 8 reset.
- Pi-compatible startup frames render header/help/context/editor/footer in that
  order and pad visible rows to terminal width.
- Resize invalidates component caches before next render.
- Renderer never writes directly to terminal; `terminal.go` owns writes.

## Diff Rules

- Diff may repaint more than strictly necessary.
- Diff must never leave stale characters on screen.
- Diff must not use append-only frame output for interactive streaming.
- Diff should prefer synchronized output plus cursor-home repaint over full
  `CSI 2J` clear-screen repaint.
- Diff must clear rows that disappeared when content shrinks.

## Input Rules

- Input decoder is byte-stream based.
- Escape sequences can be split across chunks.
- Printable UTF-8 text can be split across chunks.
- Bracketed paste is emitted as paste event, not individual key presses.
- Unknown escape sequences become raw/unknown key events, not parser panics.

## Editor Rules

- Buffer owns text, cursor, selection, undo, kill ring, and submit action.
- Rendering reads buffer state but does not mutate it.
- `Enter` submits unless keybinding maps it to newline.
- Autocomplete is a separate state machine fed by editor buffer snapshots.
- External editor receives current buffer and replaces it only after successful
  editor exit.

## Overlay Rules

- Focused visible overlay receives input first.
- Closing focused overlay restores previous focus target when still alive.
- Disposed overlay handles cannot be reused.

## Tests

- ANSI-stripped render width never exceeds terminal width;
- temporary pty byte harnesses may live outside the repository for Pi/Nu raw
  startup comparison and must not be committed;
- chunked escape sequence decodes correctly;
- bracketed paste stays one event;
- editor undo restores previous buffer/cursor;
- overlay focus stack restores previous focus.
