# `internal/tui/loop.go`

## Status

Current: IN_PROGRESS
Implementation Commit: 5d9629b
Implementation Comments: Interactive app wires stdin/stdout, renderer, editor, agent event sink, app dispatch, and is being extended to render selected model display labels.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Wire app interactive mode to the renderer, editor, stdin, stdout, and Nu agent.

## Code Style

Keep this as a small line-oriented loop until raw terminal support is required
by tests. Mark that simplification with a `ponytail:` comment. Use context-aware
agent calls and do not write outside the injected stdout/stderr.

## Types

### `type App struct`

Logic:

- Own interactive state, editor, agent pointer, IO, cwd/provider/model labels,
  optional display model label, and terminal size.
- Expose `Emit` for app-created agents.
- Expose `SetAgent` after construction.

Acceptance:

- app mode can create TUI first, then inject an agent that emits into it.

## Functions

### `NewApp(opts AppOptions) *App`

Logic:

- Normalize IO and terminal dimensions.
- Seed render state from options.
- Return without starting provider work.

Acceptance:

- construction is deterministic in tests.

### `(*App) SetAgent(a *agent.Agent)`

Logic:

- Store the agent pointer for future prompt submissions.

Acceptance:

- nil is allowed for render-only tests.

### `(*App) Emit(ev agent.Event)`

Logic:

- Update status/message state from agent events.
- Re-render the frame to stdout.
- Store write errors for `Run`.

Acceptance:

- agent events become visible UI state.

### `(*App) Run(ctx context.Context) error`

Logic:

- Render the initial frame.
- Read line-oriented input from stdin.
- Treat `/quit` and `/exit` as shutdown.
- Submit non-empty lines to the injected agent.
- Re-render after editor changes and agent events.
- Return context, scanner, prompt, or stored write errors.

Acceptance:

- `--mode interactive` is wired through app dispatch;
- tests can run the loop with fake stdin/stdout.

Tests:

- `TestNUF002DispatchInteractiveMode`
