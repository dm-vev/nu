# `internal/tui/terminal`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Small stdlib terminal wrapper for injected IO, raw mode, size queries, resize
watching, and output control sequences.

## Files

- `terminal.go`: injected stdin/stdout and dimension cache.
- `writer.go`: writes, cursor moves, cursor visibility, title.
- `constants.go`: control sequences.
- `raw_unix.go`, `raw_other.go`: raw mode.
- `size_unix.go`, `size_other.go`: terminal dimensions.
- `resize_unix.go`, `resize_other.go`: resize watcher.
- `terminal_test.go`: write-error behavior.

## Acceptance Criteria

- Raw mode returns a restore callback only after successful setup.
- Tests use injected IO and do not require the developer terminal.
- Platform-specific code stays behind build tags.
