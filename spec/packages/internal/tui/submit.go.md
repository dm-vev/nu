# `internal/tui/tui_submit.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Prompt submission, local slash command dispatch, Pi-style model selector wiring, prompt error surfacing, turn aborts, and render error storage.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Prompt submission, prompt error surfacing, turn aborts, and render error storage.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui`.
## Functions

### `func (a *App) submit(value string) error`

Logic:
- Trim empty input.
- Dispatch slash commands locally before appending user messages.
- Append ordinary user message, render immediately, and start provider prompt asynchronously.

Acceptance:
- Empty input does not call the agent; visible user text appears before provider streaming starts.
- Slash commands do not reach the model as normal prompts.

### `func (a *App) runSlashCommand(name string, args string) error`

Logic:
- Reject unknown commands with a local message.
- Dispatch every Pi builtin slash command to a local handler.
- Keep unknown commands local instead of forwarding them as prompts.
- `/new` clears both visible TUI messages and the agent's remembered provider history.

Acceptance:
- `TestTUISlashSessionDoesNotCallAgent` fails if `/session` calls the agent.
- `TestTUIAllBuiltinSlashCommandsHaveHandlers` fails if any builtin falls back to a placeholder.

### `func (a *App) appendLocalMessage(text string)`

Logic:
- Append a local assistant message and render.

Acceptance:
- Slash command output appears in chat.

### `func (a *App) sessionCommandText() string`

Logic:
- Build a small Markdown table with cwd, provider, model, and message count.

Acceptance:
- `/session` displays useful local state.

### `func (a *App) handleModelSlash(args string) error`

Logic:
- Without args, open the model selector.
- With args, select an exact model match if possible.
- Otherwise open the selector with the argument as its search query.

Acceptance:
- `TestTUISlashModelOpensSelectorAndSelectsModel` fails if `/model` does not open the selector.
- `TestTUISlashModelExactMatchSelectsWithoutMenu` fails if exact model references do not select immediately.

### `func (a *App) openModelMenu(query string)`

Logic:
- Refresh selector candidates, open it with current provider/model id, clear transient status, and render.

Acceptance:
- `/model` opens a selector without appending a chat message.

### `func (a *App) findModel(query string) (model.Model, bool)`

Logic:
- Resolve a query against visible model candidates using the model registry matching rules.

Acceptance:
- Aliases and provider-qualified ids can select a model without opening the menu.

### `func authAll(models []model.Model) map[string]bool`

Logic:
- Build a permissive auth map for already-filtered visible model candidates.

Acceptance:
- Exact matching does not hide entries that runtime already proved visible.

### `func (a *App) selectModel(selected model.Model) error`

Logic:
- Validate selected provider/api/id.
- Call the runtime model callback when present, otherwise update the injected agent labels.
- Update raw model id, display label, context or default context, footer, and close the selector.
- Clear transient status and append `Model switched from ... to ...` as a normal chat message.

Acceptance:
- Selecting from `/model` updates both footer and the provider-backed agent path without starting the status spinner.

### `func modelDisplayName(entry model.Model) string`

Logic:
- Prefer custom display name, falling back to raw model id.

Acceptance:
- Custom model display names are shown in footer/status.

### `func (a *App) handleModelMenuInput(data string) bool`

Logic:
- Route raw input into the visible selector.
- Cancel on selector cancel action, select on confirm, and surface selection errors in chat.

Acceptance:
- Arrow/Enter/Escape input is not forwarded to the editor while the model selector is visible.

### `func (a *App) startPrompt(agentRef *agent.Agent, value string)`

Logic:
- Run agent prompt in a goroutine and convert prompt errors into visible assistant error messages.

Acceptance:
- The wait group is incremented before launching and decremented exactly once when the prompt returns.

### `func (a *App) appendError(err error)`

Logic:
- Clear status, append a visible assistant error, rebuild chat, and render.

Acceptance:
- Prompt errors are visible in the chat instead of being written to stdout/stderr directly.

### `func (a *App) abortActiveTurn() bool`

Logic:
- Abort a busy agent and show an aborting status without exiting the UI.

Acceptance:
- Returns false without side effects when no agent is busy.

### `func (a *App) rememberWriteErr(err error)`

Logic:
- Store the first render/write error so `Run` can return it after input exits.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (a *App) render()`

Logic:
- Delegate rendering to `TUI.RenderNow` and remember the first render error.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
