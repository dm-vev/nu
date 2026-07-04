# `internal/tui/layout.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Build the fixed component tree and component options, including model selector rows above command suggestions.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Build the fixed component tree and component options.

## Code Style

Keep this file small, stdlib-only, and covered by narrow tests. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Functions

### `func (a *App) buildLayout()`

Logic:
- Add header, chat, flexible fill, command menu, status, editor, and footer to the engine in fixed Pi-style order.
- Add model selector rows above slash command suggestions so `/model` owns input without entering chat scrollback.
- Keep status directly above the editor so streaming state never appears inside the scrollback area.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func headerOptions(opts AppOptions) header.Options`

Logic:
- Build header options with Nu label, version, palette callbacks, and one-cell horizontal padding.
- Use ` | ` as the compact-help separator in limited-character mode, otherwise keep the Pi-style ` · ` separator.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func footerOptions(opts AppOptions) footer.Options`

Logic:
- Build footer options from cwd/home/branch/provider/model display label/context and dim styling.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func commandMenuOptions() commandmenu.Options`

Logic:
- Build command menu options with Pi-style row cap and Nu palette callbacks.

Acceptance:
- Command menu renders without touching editor or agent state.

### `func modelMenuOptions() modelmenu.Options`

Logic:
- Build model selector options with Pi-style row cap and Nu palette callbacks.

Acceptance:
- Model selector renders without touching editor or agent state.
