# `internal/tui/editor_editor.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Define editor state owner and public editor methods.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define editor state owner and public editor methods.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Editor struct {`

Logic:
- Editor is a focused input component.
- It stores an optional border rune so limited terminals can render an ASCII prompt line.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func New() *Editor`

Logic:
- New creates an editor component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) SetSubmitHandler(handler func(string))`

Logic:
- SetSubmitHandler sets the submit callback.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) SetChangeHandler(handler func(string))`

Logic:
- SetChangeHandler sets the change callback.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) SetStyles(border func(string) string, textStyle func(string) string)`

Logic:
- SetStyles sets border and text styles.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) SetBorderRune(borderRune rune)`

Logic:
- Set the prompt border glyph used by `Render`.

Acceptance:
- `TestEditorEditorCanUseASCIIBorder` fails if ASCII prompt lines cannot be selected.

### `func (e *Editor) SetFocused(focused bool)`

Logic:
- SetFocused updates focus state.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) Text() string`

Logic:
- Text returns editor text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) Clear()`

Logic:
- Clear resets the editor buffer.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (e *Editor) SetText(value string)`

Logic:
- Replace the editor with one logical line and move the cursor to the end.

Acceptance:
- Command menu completion can atomically replace the current slash prefix.
