# `internal/tui/app.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define the interactive App state, model selector state, and construction boundary.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define the interactive App state and construction boundary.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.

## Types

### `type App struct {`

Logic:
- Own the top-level interactive state: agent pointer, terminal, engine, editor,
  header/chat/command menu/model selector/status/footer components, cwd/provider/model metadata, structured
  `internal/tui` chat state, pending prompt wait group, and last write
  error.
- Track raw model id separately from the display label so footer text can use custom names while provider requests use real ids.
- Track session id/name labels for local `/session` and `/name`.

Acceptance:
- Reusable component, editor, terminal, and renderer logic stays in its
  approved TUI subpackages, with all reusable components in `tui/components`.

## Functions

### `func NewApp(opts AppOptions) *App`

Logic:
- NewApp creates an idle interactive app.
- Wire editor change events into the command menu.
- Build model selector candidates from explicit model metadata, falling back to the current model when no registry list is supplied.
- Apply limited-character terminal settings by replacing Unicode status frames with `-\|/`, rendering the prompt border as a muted ASCII line, and replacing compact-header Unicode separators.

Acceptance:
- Construction performs no provider work and can be tested with injected stdin/stdout.

### `func modelChoices(opts AppOptions) []model.Model`

Logic:
- Return copied model candidates when supplied.
- Fall back to a single current model entry for injected-provider tests and compatibility paths.

Acceptance:
- `/model` can open even in tests that inject only current provider/model labels.

### `func (a *App) SetAgent(agentRef *agent.Agent)`

Logic:
- SetAgent injects the provider-backed agent.

Acceptance:
- Nil is allowed for render-only tests; non-nil agents are used only by submit paths.

### `func (a *App) requestQuit()`

Logic:
- Mark the interactive app for loop exit.

Acceptance:
- `/quit` can leave raw mode through normal cleanup.

### `func (a *App) shouldQuit() bool`

Logic:
- Report whether slash command handling requested exit.

Acceptance:
- Raw and line loops can stop after command submission.
