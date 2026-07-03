# `internal/tui/loop.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Interactive app wires stdin/stdout, raw TTY input fallback, renderer, editor, terminal writer, agent event sink, app dispatch, selected model display labels, version/home/context metadata, and in-place repaint output.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Wire app interactive mode to the renderer, editor, stdin, stdout, and Nu agent.

## Code Style

Keep raw terminal support narrow and stdlib-only. Use context-aware agent calls,
restore terminal state on exit, and do not write outside injected stdout/stderr.

## Types

### `type App struct`

Logic:

- Own interactive state, editor, agent pointer, IO, cwd/home/branch,
  provider/model labels, optional display model label, version/context metadata,
  terminal writer, and terminal size.
- Expose `Emit` for app-created agents.
- Expose `SetAgent` after construction.

Acceptance:

- app mode can create TUI first, then inject an agent that emits into it.

## Functions

### `NewApp(opts AppOptions) *App`

Logic:

- Normalize IO and terminal dimensions, using `COLUMNS`/`LINES` before fallback
  defaults.
- Seed render state from options.
- Create a repainting terminal writer for interactive mode.
- Read `.git/HEAD` directly for branch footer when no branch is supplied.
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

- Enable raw input when stdin is a TTY; otherwise use deterministic line mode
  for tests and pipes.
- Render the initial frame and close/restore terminal state on return.
- In raw mode, handle printable runes, backspace, enter submit, Ctrl-D empty
  exit, and Ctrl-C exit.
- In line mode, treat `/quit` and `/exit` as shutdown.
- Submit non-empty input to the injected agent.
- Re-render after editor changes and agent events.
- Return context, scanner, prompt, or stored write errors.

Acceptance:

- `--mode interactive` is wired through app dispatch;
- tests can run the loop with fake stdin/stdout.
- repeated streaming updates repaint one terminal frame instead of appending
  duplicate frames.

Tests:

- `TestNUF002DispatchInteractiveMode`
- `TestTUIAppRenderUsesTerminalRepaint`
