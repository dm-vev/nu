# `internal/tui/ansi`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

ANSI and terminal-cell utilities. Components use this package to wrap,
truncate, pad, strip, and measure text without corrupting SGR/OSC/APC escape
sequences.

## Files

- `escape.go`: escape sequence boundary parsing and reset suffix.
- `style.go`: built-in SGR color constants.
- `width.go`: visible cell width calculation.
- `strip.go`: escape removal.
- `truncate.go`: ANSI-preserving width slicing.
- `wrap.go`: ANSI-preserving word wrapping.
- `pad.go`: right padding/truncation to exact width.
- `ansi_test.go`: width, wrap, and escape behavior checks.

## Acceptance Criteria

- Escape bytes never count as visible terminal cells.
- Width-bound helpers never return ANSI-stripped text wider than requested.
- Package has no component or terminal writer dependencies.
