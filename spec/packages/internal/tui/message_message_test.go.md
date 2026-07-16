# `internal/tui/message_message_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Tests structured message mutation invariants.

## TODO

- [x] Tests exist for coalescing, ordering, and replace behavior.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Prevent regressions where TUI message parts collapse back into one string.

## Acceptance Criteria

- `go test ./internal/tui` passes.

## Tests

### `TestMessageAppendTextCoalescesAdjacentTextParts`

Logic:
- Append two text deltas and assert one text part remains.

Acceptance:
- Adjacent streamed text does not create unnecessary render blocks.

### `TestMessageThinkingAndToolPartsKeepOrdering`

Logic:
- Append thinking, tool, and text parts and assert order/state.

Acceptance:
- Mixed assistant content keeps Pi-like ordering.

### `TestMessageReplaceTextDoesNotDeleteToolParts`

Logic:
- Replace final text in a message that already has a tool part.

Acceptance:
- Tool blocks survive final text replacement.
