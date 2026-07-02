# `internal/tui/input.go`

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

Decode terminal input bytes into key events, paste events, resize, and signals.
Implements input rules from `spec/protocols/tui-rendering.md`.

## Code Style

Parser is stateful but deterministic. Chunk boundary tests are required.

## Functions

### `DecodeInput(dec *Decoder, data []byte) ([]InputEvent, error)`

Logic:

- Append incoming bytes to decoder buffer.
- Decode complete UTF-8 runes into text events when outside escape/paste modes.
- Detect ESC and wait for enough bytes to classify known CSI/SS3/Kitty
  sequences; keep incomplete sequences buffered.
- Decode bracketed paste start/end and emit one paste event with raw text.
- Map known control bytes and escape sequences to normalized key events.
- Emit unknown complete escape sequences as unknown-key events, not errors.
- Return parse error only for invalid internal decoder state.

Acceptance:

- decodes printable text, control keys, arrows, modifiers, paste brackets, and
  Kitty key protocol where supported.
- handles chunked UTF-8 and chunked escape sequences.

Tests:

- `TestTUIInputDecodeChunkedEscape`
