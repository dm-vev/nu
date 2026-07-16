# `internal/tui/components/assistant_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Tests assistant text and structured part rendering.

## TODO

- [x] Test file is runnable with `go test ./internal/tui`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Lock assistant message behavior against regressions back to string-only output.

## Tests

### `TestAssistantMessageAssistantMessageRendersTextAndZone`

Logic:
- Render a text-only assistant message.

Acceptance:
- Output contains text and starts with OSC 133 prompt-zone marker.

### `TestAssistantMessageAssistantMessageRendersPartsWithoutZoneWhenToolExists`

Logic:
- Render thinking, tool, and final text parts in one assistant message.

Acceptance:
- Output contains all visible parts and does not wrap the whole message in OSC markers.
