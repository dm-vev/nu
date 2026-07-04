# `internal/tui/app_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Verify top-level TUI app rendering, slash command menus, model selector behavior, and raw input behavior.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Test file is runnable with `go test ./internal/tui/...`.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Verify top-level TUI app rendering and raw input behavior.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Test remains narrow and does not require real providers or real `~/.nu`.
## Functions

### `func TestTUIAppRendersPiStyleComponentTree(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestTUIHandleRawInputTogglesHeader(t *testing.T)`

Logic:
- Exercise the behavior named by the test and fail on visible regressions.

Acceptance:
- The test fails if the named behavior regresses.

### `func TestTUIAppKeepsStatusLineAboveEditor(t *testing.T)`

Logic:
- Render an active turn and assert the status row sits immediately above the editor border.

Acceptance:
- The test fails if status returns to the chat area or collapses in idle/busy layout.

### `func TestTUIAppUsesLimitedCharsetWhenRequested(t *testing.T)`

Logic:
- Construct the app with `ASCII: true`, start a turn, and assert the rendered TUI uses ASCII spinner frames and muted ASCII prompt lines.

Acceptance:
- The test fails if unsupported Unicode status or prompt glyphs appear in limited-character mode.

### `func editorBorderLine(line string) bool`

Logic:
- Return true when a rendered line contains either the Unicode or ASCII editor border glyph.

Acceptance:
- Layout tests stay focused on editor anchoring instead of duplicating charset fallback assertions.

### `func TestTUIHandleRawInputScrollsViewport(t *testing.T)`

Logic:
- Build overflowing chat history, render the bottom viewport, send PageUp, and assert the viewport changes to older content.

Acceptance:
- The test fails if PageUp is forwarded to the editor or if manual scroll does not affect rendering.

### `func TestTUIRateLimitShowsFooterNotice(t *testing.T)`

Logic:
- Emit a `rate_limit` event and assert `Rate limit` appears on the footer path row with red styling.

Acceptance:
- The test fails if rate-limit retry state is rendered only in chat/status or loses the red notice style.

### `func TestTUICommandMenuRendersAndCompletes(t *testing.T)`

Logic:
- Type `/mo`, assert command suggestions render, press Tab, and assert editor contains `/model `.

Acceptance:
- The test fails if command menu filtering or completion breaks.

### `func TestTUICommandMenuEnterRunsSelectedCommand(t *testing.T)`

Logic:
- Type `/`, move the command selector to `/model`, press Enter, and assert the model selector opens.

Acceptance:
- The test fails if Enter submits raw slash text or ignores the highlighted command row.

### `func TestTUISlashSessionDoesNotCallAgent(t *testing.T)`

Logic:
- Submit `/session` without an agent and assert local session output renders.

Acceptance:
- The test fails if slash commands are routed through provider prompt handling.

### `func TestTUISlashQuitRequestsExit(t *testing.T)`

Logic:
- Submit `/quit` and assert the app requests loop exit.

Acceptance:
- The test fails if raw mode cannot exit through the Pi slash command.

### `func TestTUISlashModelOpensSelectorAndSelectsModel(t *testing.T)`

Logic:
- Submit `/model`, assert the selector renders, move down, press Enter, and assert model selection writes a chat notice and updates the footer.

Acceptance:
- The test fails if `/model` only prints text, if selector confirmation does not update the selected model, or if the selected model is rendered in animated status.

### `func TestTUISlashModelExactMatchSelectsWithoutMenu(t *testing.T)`

Logic:
- Submit `/model provider/id` and assert the model callback runs without leaving the selector open.

Acceptance:
- The test fails if exact model references always open the menu or do not select.

### `func TestTUIAllBuiltinSlashCommandsHaveHandlers(t *testing.T)`

Logic:
- Submit every command from `slash.Builtins()` with safe arguments where needed.
- Assert none of them render the old backend placeholder.

Acceptance:
- The test fails if a copied Pi menu item has no local TUI handler.
