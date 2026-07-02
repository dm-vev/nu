# `internal/tui/editor.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 5d9629b
Implementation Comments: Rune-safe editor buffer supports insert, backspace, clamped movement, undo snapshots, submit, and immutable snapshots without renderer or terminal coupling.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Own deterministic editor buffer mutations for interactive mode.

## Code Style

Use rune slices for cursor-safe text editing. Keep rendering out of the editor.
Comment undo snapshot decisions and submit semantics.

## Types

### `type Editor struct`

Logic:

- Store text as runes, cursor index, undo snapshots, and last submitted text.
- Keep selection, kill ring, autocomplete, and external-editor hooks as explicit
  fields or future state boundaries, not hidden renderer behavior.

Acceptance:

- edits are UTF-8 safe;
- rendering can read snapshots without mutating the buffer.

### `type EditorSnapshot struct`

Logic:

- Carry text, cursor, and submitted value for renderer/tests.

Acceptance:

- callers can assert state without peeking at private fields.

## Functions

### `NewEditor() *Editor`

Logic:

- Return an empty editor with cursor at zero.

Acceptance:

- no terminal or agent dependency.

### `(*Editor) Insert(text string)`

Logic:

- Save undo state.
- Insert runes at the current cursor.
- Advance cursor by inserted rune count.

Acceptance:

- supports multi-rune and multi-line input.

### `(*Editor) Backspace()`

Logic:

- If cursor is at zero, do nothing.
- Save undo state.
- Remove the rune before cursor and move cursor back one.

Acceptance:

- delete never panics at buffer boundaries.

### `(*Editor) Move(delta int)`

Logic:

- Clamp cursor movement to `[0, len(buffer)]`.

Acceptance:

- cursor remains valid after any move.

### `(*Editor) Undo() bool`

Logic:

- Restore the previous buffer/cursor snapshot when present.
- Return false when there is nothing to undo.

Acceptance:

- undo restores text and cursor together.

### `(*Editor) Submit() string`

Logic:

- Return current text.
- Store it as last submitted text.
- Clear buffer and cursor only after capturing the text.

Acceptance:

- submitted text is exact;
- subsequent editing starts from an empty buffer.

### `(*Editor) Snapshot() EditorSnapshot`

Logic:

- Return a copy of the current visible state.

Acceptance:

- callers cannot mutate editor internals through the snapshot.

Tests:

- `TestNUF101EditorInsertDeleteUndo`
