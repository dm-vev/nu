# `internal/tui/terminal.go`

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

Terminal abstraction for raw mode, size, writes, and lifecycle.

## Code Style

Platform-specific syscalls stay in `internal/platform`. This package consumes a
terminal interface.

## Functions

### `New(term Terminal, opts Options) *TUI`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Do not enter raw mode until `Start`.

Acceptance:

- does not enter raw mode until `Start`.

### `(*TUI) Start(ctx context.Context) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Initialize local state, then enter the smallest required loop.
- Stop on context cancellation, terminal command, or unrecoverable error and clean up owned resources.
- Enters raw mode.
- Start input/render loops.
- Restores terminal on exit.

Acceptance:

- enters raw mode;
- starts input/render loops;
- restores terminal on exit.

### `(*TUI) Stop() error`

Logic:

- Mark the TUI as stopping so render and input loops exit once they observe cancellation.
- Call the raw-mode restore function exactly once.
- Flush any final terminal reset/write operations needed to leave the screen usable.
- Return the first restore/flush error and ignore repeated stop calls.

Acceptance:

- idempotently restores terminal.

Tests:

- `TestTUIStartStopRestoresTerminal`
