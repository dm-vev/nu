# `internal/tui/input`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Raw terminal input decoder. It groups UTF-8 runes, escape sequences, and
bracketed paste payloads before top-level app key handling sees them.

## Files

- `event.go`: decoded event type and paste markers.
- `decoder.go`: stream decoder for UTF-8 and escape sequences.
- `paste.go`: bracketed paste payload reader.
- `decoder_test.go`: UTF-8, arrow, EOF, and paste checks.

## Acceptance Criteria

- Printable Unicode is emitted as complete strings.
- Arrow/delete/paste escape sequences are not split into individual bytes.
- The decoder has no editor or app dependencies.
