# `internal/testkit/terminal.go`

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

Fake terminal for TUI tests.

## Code Style

Deterministic frame capture. No real tty.

## Functions

### `NewTerminal(size tui.Size) *Terminal`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Supply scripted input bytes.
- Record writes and raw-mode lifecycle.
- Support resize events.

Acceptance:

- supplies scripted input bytes;
- records writes and raw-mode lifecycle;
- supports resize events.

Tests:

- used by `internal/tui` tests.
