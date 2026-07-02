# `internal/tui/editor.go`

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

Interactive multiline editor.
Implements editor rules from `spec/protocols/tui-rendering.md`.

## Code Style

Separate text buffer operations from rendering. Buffer operations are pure enough
for unit tests.

## Functions

### `NewEditor(opts EditorOptions) *Editor`

Logic:

- Initialize text buffer as one empty line, cursor at row 0 col 0.
- Install keymap reference and autocomplete providers from options.
- Initialize undo stack with empty baseline and empty kill ring.
- Apply editor padding, placeholder, and multiline settings defaults.
- Do not inspect filesystem until autocomplete is requested.

Acceptance:

- initializes empty buffer, cursor, autocomplete state, and history hooks.

### `(*Editor) Handle(event InputEvent) EditorAction`

Logic:

- Resolve input event to action id through keybindings.
- If autocomplete menu is open, route navigation/confirm/cancel to autocomplete
  first.
- For text insertion, push undo boundary, insert at cursor, advance cursor by
  grapheme/rune policy defined by editor tests.
- For movement/deletion, update buffer and cursor while preserving valid cursor
  position.
- For kill-ring actions, store deleted spans and support yank/yank-pop.
- For submit, return `EditorActionSubmit` with current buffer and do not clear
  buffer until caller accepts.
- For external editor, return action with temp-file request; caller performs
  process IO and feeds replacement text back.
- For paste image, return action that app layer resolves through platform
  clipboard/image handling.

Acceptance:

- supports movement, deletion, undo, kill ring, newline, submit, autocomplete,
  external editor, and paste image actions.

Tests:

- `TestNUF101EditorInsertDeleteUndo`
- `TestNUF101AutocompleteAtFileReference`
