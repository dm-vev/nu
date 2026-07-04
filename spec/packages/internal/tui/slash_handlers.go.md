# `internal/tui/slash_handlers.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Local handlers for all built-in Pi slash commands that can run against current TUI state without adding a new session service.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by package-level TUI tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Keep slash command behavior out of `submit.go`. Handlers either mutate current
TUI state, read/write the current in-memory chat through simple files, or update
the global auth/trust files already used by runtime wiring.

## Code Style

Use stdlib only. Do not call real providers. Do not silently swallow command
errors that affect files. Keep deliberate local-only behavior marked with a
`ponytail:` comment when it has a known ceiling.

## Acceptance Criteria

- No built-in command renders the old backend placeholder.
- Tests do not read or write real `~/.nu`; they pass injected temp `Home`/`CWD`.
- File-writing commands wrap path errors with context.

## Types

### `type slashExportRecord`

Logic:
- Store one exported JSONL message with role and plain text.

Acceptance:
- `/export *.jsonl`, `/import`, and `/resume` share one minimal interchange format.

### `type modelSummary`

Logic:
- Store display data for `/scoped-models`.

Acceptance:
- Rendering model scope does not expose full provider metadata.

## Functions

### `func (a *App) settingsCommandText() string`

Logic:
- Render current cwd, home, provider, model, context, and session label as a Markdown table.

Acceptance:
- `/settings` is local and side-effect free.

### `func (a *App) scopedModelsCommandText() string`

Logic:
- Render visible model candidates with the active model marked.

Acceptance:
- `/scoped-models` reports the model list the selector can use.

### `func (a *App) handleExportSlash(args string) error`

Logic:
- Export current in-memory chat to a requested path or `nu-session.html`.

Acceptance:
- JSONL, Markdown, and HTML export paths write files.

### `func (a *App) handleImportSlash(args string, replace bool) error`

Logic:
- Import exported JSONL messages.
- Append for `/import`; replace current messages for `/resume`.

Acceptance:
- Missing path shows usage; invalid files return wrapped errors.

### `func (a *App) handleShareSlash(args string) error`

Logic:
- Write a share-ready HTML export to the requested path or temp dir.

Acceptance:
- `/share` has a concrete artifact without doing network work.

### `func (a *App) handleCopySlash() error`

Logic:
- Copy the last assistant text through a native clipboard command when present.
- Render a local failure message when no clipboard tool exists.

Acceptance:
- `/copy` never reaches the model.

### `func (a *App) handleNameSlash(args string)`

Logic:
- Show or update the current session name.

Acceptance:
- `/session` reflects names set by `/name`.

### `func (a *App) handleChangelogSlash() error`

Logic:
- Render `CHANGELOG.md` or `docs/changelog.md` when present, else report absence.

Acceptance:
- `/changelog` is a local file read.

### `func (a *App) handleForkSlash(args string)`

Logic:
- Keep messages through a requested 1-based message index or the latest user message.

Acceptance:
- `/fork` creates a local in-memory branch by truncating later messages.

### `func (a *App) handleCloneSlash()`

Logic:
- Clone current message values in memory and mark status.

Acceptance:
- The command has no provider or filesystem dependency.

### `func (a *App) treeCommandText() string`

Logic:
- Render a linear message tree with index, role, and preview.

Acceptance:
- `/tree` gives navigable message indexes for `/fork N`.

### `func (a *App) handleTrustSlash(args string) error`

Logic:
- Save current cwd trust to `~/.nu/agent/trust.json` or injected home equivalent.

Acceptance:
- `/trust` writes the same global trust surface described by product spec.

### `func (a *App) handleLoginSlash(args string) error`

Logic:
- Save provider auth as direct key, env reference, or command reference in `~/.nu/auth.json`.

Acceptance:
- `/login provider env NAME` can configure credentials without echoing a secret.

### `func (a *App) handleLogoutSlash(args string) error`

Logic:
- Remove one provider from `~/.nu/auth.json`.

Acceptance:
- `/logout provider` leaves other providers intact.

### `func (a *App) handleCompactSlash()`

Logic:
- Locally reduce long in-memory history to a compact marker plus recent tail.

Acceptance:
- Small histories report that nothing needs compacting.

### `func (a *App) handleReloadSlash()`

Logic:
- Refresh git branch footer state and show reloaded status.

Acceptance:
- `/reload` has a visible local effect.

### `func (a *App) sessionSnapshot() []message.Message`

Logic:
- Clone current messages under lock for export.

Acceptance:
- Export does not race with TUI event updates.

### `func (a *App) exportMessages(path string) error`

Logic:
- Pick JSONL, Markdown, or HTML rendering by path extension and write mode `0600`.

Acceptance:
- Export errors include the target path.

### `func (a *App) resolvePath(path string) string`

Logic:
- Resolve relative paths under current cwd.

Acceptance:
- Tests can keep all slash file IO inside temp dirs.

### `func (a *App) authPath() string`

Logic:
- Return the auth file path under injected home, falling back to cwd.

Acceptance:
- `/login` and `/logout` do not read real home in tests.

### `func (a *App) trustPath() string`

Logic:
- Return the trust file path under injected home, falling back to cwd.

Acceptance:
- `/trust` does not read real home in tests.

### `func (a *App) lastAssistantText() string`

Logic:
- Find the latest assistant message text from newest to oldest.

Acceptance:
- `/copy` ignores user-only history.

### `func (a *App) forkIndexLocked(args string) int`

Logic:
- Parse a 1-based message index or find the latest user message.

Acceptance:
- `/fork N` and bare `/fork` share one selection path.

### `func summarizeModels(models []model.Model, providerID string, modelID string) []modelSummary`

Logic:
- Convert model metadata into table rows and mark the current model.

Acceptance:
- `/scoped-models` stays display-only.

### `func readExportedMessages(path string) ([]message.Message, error)`

Logic:
- Decode JSONL export records into TUI messages.

Acceptance:
- Unknown roles and malformed JSON fail clearly.

### `func renderJSONL(messages []message.Message) ([]byte, error)`

Logic:
- Encode role/text records as JSONL.

Acceptance:
- Exported JSONL can be imported by the same code.

### `func renderMarkdown(messages []message.Message) string`

Logic:
- Render each message as a Markdown section.

Acceptance:
- `.md` exports are readable without the TUI.

### `func renderHTML(messages []message.Message) string`

Logic:
- Render escaped message text in a standalone HTML document.

Acceptance:
- Default export/share produces inspectable HTML.

### `func messageText(msg message.Message) string`

Logic:
- Concatenate text and tool-visible fields from one message.

Acceptance:
- Export/copy/tree preview use the same plain text.

### `func copyToClipboard(text string) error`

Logic:
- Try `wl-copy`, `xclip`, `xsel`, then `pbcopy`.

Acceptance:
- Missing clipboard support returns a visible local error.

### `func writeBoolMap(path string, key string, value bool) error`

Logic:
- Merge and write a JSON boolean map.

Acceptance:
- `/trust no` can update an existing trust file.

### `func writeAuthProvider(path string, providerID string, credential map[string]string) error`

Logic:
- Merge one provider credential into auth JSON.

Acceptance:
- `/login` does not delete other providers.

### `func removeAuthProvider(path string, providerID string) error`

Logic:
- Delete one provider credential from auth JSON.

Acceptance:
- `/logout` does not delete other providers.

### `func writeJSONFile(path string, value any) error`

Logic:
- Create parent directories, marshal indented JSON, and write mode `0600`.

Acceptance:
- File write errors include path context.

### `func tableCell(value string) string`

Logic:
- Escape Markdown table pipes and collapse newlines.

Acceptance:
- `/tree` previews do not break table layout.

### `func preview(value string, limit int) string`

Logic:
- Normalize whitespace and truncate long rune slices with ellipsis.

Acceptance:
- `/tree` remains compact for long messages.
