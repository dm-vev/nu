# `internal/tui/input.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 5d9629b
Implementation Comments: Byte-stream decoder handles split UTF-8, split escape sequences, common control keys, arrow keys, bracketed paste as one event, unknown escapes, and EOF flushing without terminal dependencies.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Decode terminal byte streams into deterministic key/text/paste events.

## Code Style

Use `unicode/utf8` for split rune handling. Avoid terminal dependencies here.
Comment incomplete escape/paste buffering.

## Types

### `type InputEvent struct`

Logic:

- Carry kind, key name, text payload, and raw sequence.

Acceptance:

- tests can compare decoded events without terminal state.

### `type Decoder struct`

Logic:

- Store pending bytes across chunks.
- Recognize common control keys, arrow escape sequences, unknown escape
  sequences, printable UTF-8 text, and bracketed paste.

Acceptance:

- chunk boundaries do not change decoded events.

## Functions

### `NewDecoder() *Decoder`

Logic:

- Return a decoder with empty pending buffer.

Acceptance:

- no terminal side effects.

### `(*Decoder) Write(chunk []byte) []InputEvent`

Logic:

- Append chunk to pending bytes.
- Emit complete printable text runs as text events.
- Emit bracketed paste content as one paste event.
- Hold incomplete UTF-8, escape, and paste sequences for the next chunk.
- Emit unknown complete escape sequences as unknown events.

Acceptance:

- split escape sequences decode correctly;
- bracketed paste is one event;
- malformed input does not panic.

### `(*Decoder) Flush() []InputEvent`

Logic:

- Emit any pending bytes as text or unknown events.
- Clear the buffer.

Acceptance:

- EOF does not lose pending input.

Tests:

- `TestNUF100InputDecodesChunkedEscape`
- `TestNUF100BracketedPasteIsSingleEvent`
