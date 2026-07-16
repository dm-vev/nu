# `internal/app/eventoutput.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Print and JSON event sinks now have a dedicated output-only file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Translate agent lifecycle events into print-mode text or JSON-mode JSONL.

## Code Style

Keep protocol stdout exclusive and retain the first output error.

## Owned Logic

- `printEventWriter.emit` writes only non-empty final `turn_end` assistant text.
- `jsonEventWriter.emit` writes every event as one JSONL record.

## Acceptance

- Print mode does not expose live deltas or diagnostics.
- JSON mode emits only valid JSONL and stops writing after failure.

## Tests

- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
