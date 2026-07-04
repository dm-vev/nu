# `internal/tui/run.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Own raw/line input loops and terminal lifecycle.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Own raw/line input loops and terminal lifecycle.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.
## Functions

### `func (a *App) Run(ctx context.Context) (runErr error)`

Logic:
- Run starts the interactive loop.
- It installs the editor submit handler, enables raw mode when possible, starts the render engine, renders the initial frame, runs raw or line input, waits for active prompt goroutines, then stops the engine and restores raw mode.
- It starts a small status ticker while the app is running so active state labels animate.

Acceptance:
- Terminal state is restored or cleanup is returned on every successful setup path.
- Prompt goroutines cannot outlive `ui.Stop()` during normal shutdown.

### `func (a *App) runLine(ctx context.Context) error`

Logic:
- Read newline-delimited input for pipes/tests, honor `/quit` and `/exit`, and submit non-empty lines.
- Stop when slash command handling requests quit.

Acceptance:
- Scanner/context/prompt errors are returned; ordinary submitted lines are rendered before provider work completes.

### `func (a *App) runRaw(ctx context.Context) error`

Logic:
- Read decoded terminal events until EOF/context cancellation, apply raw input actions, and render after each event.
- Stop when slash command handling requests quit.

Acceptance:
- Decoder errors are wrapped with `read tui input`; EOF returns the stored write error.

### `func (a *App) handleRawInput(data string) bool`

Logic:
- Handle exit/abort/clear/header toggle shortcuts before forwarding ordinary input to the editor.
- Let an open model selector consume raw input before global shortcuts/editor input.
- Let an open command selector consume up/down/Enter before editor input.
- Handle PageUp/PageDown/End and mouse wheel sequences as viewport scroll actions.
- Use Tab to complete the highlighted slash command menu entry.

Acceptance:
- Ctrl+C/Esc abort a busy agent or clear the editor; Ctrl+D exits only when the editor is empty and the agent is idle; Ctrl+O toggles header expansion.
- PageUp/PageDown and mouse wheel input scroll the viewport without mutating editor text.
- Tab completion updates editor text only when a command suggestion is visible.
- Enter on a visible command suggestion executes the highlighted command instead of submitting raw `/` text.
- Arrow/Enter/Escape input drives the model selector while it is visible.

### `func (a *App) completeCommand() bool`

Logic:
- Complete the highlighted visible command menu item into the editor.

Acceptance:
- `TestTUICommandMenuRendersAndCompletes` fails if Tab completion stops working.

### `func (a *App) handleCommandMenuInput(data string) bool`

Logic:
- Route up/down arrows to command-menu selection.
- Route Enter to the highlighted slash command and clear the editor.
- Surface handler errors as visible assistant errors.

Acceptance:
- `TestTUICommandMenuEnterRunsSelectedCommand` fails if Enter submits raw slash text.

### `func isWheelUp(data string) bool`

Logic:
- Recognize SGR and legacy X10 mouse wheel-up escape sequences.

Acceptance:
- Wheel-up events route to `ScrollBy` instead of editor input.

### `func isWheelDown(data string) bool`

Logic:
- Recognize SGR and legacy X10 mouse wheel-down escape sequences.

Acceptance:
- Wheel-down events route to `ScrollBy` instead of editor input.

### `func (a *App) startStatusTicker(ctx context.Context) func()`

Logic:
- Start a ticker that advances and renders the status animation only while status text is non-empty.
- Return a stop function used by `Run` cleanup.

Acceptance:
- Ticker stops on app context cancellation or returned stop function.
