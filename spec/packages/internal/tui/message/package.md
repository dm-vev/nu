# `internal/tui/message`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Structured chat message state for the TUI. This package keeps user text,
assistant text, model thinking, and tool execution blocks as ordered parts so
rendering components do not infer structure from a single concatenated string.

## Files

- `roles.go`: exported role constants.
- `part.go`: message part and tool state types.
- `message.go`: message constructors and clone helper.
- `mutate.go`: append/replace/finalize mutation helpers.
- `message_test.go`: ordering and mutation tests.

## Acceptance Criteria

- Message parts preserve provider/render order.
- Adjacent text and thinking deltas coalesce only with the same part kind.
- Tool blocks can be appended pending and later finalized by id.
- `go test ./internal/tui/message` passes.
