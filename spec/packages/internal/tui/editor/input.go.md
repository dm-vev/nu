# `internal/tui/editor/input.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Mutate editor buffer from decoded key/input sequences.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Mutate editor buffer from decoded key/input sequences.

## Code Style

Mutate by rune positions, not bytes. Keep input mutation separate from rendering. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `const (...)`

Logic:
- Define bracketed paste start/end markers recognized by `HandleInput`.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func (e *Editor) HandleInput(data string)`

Logic:
- HandleInput updates editor state from raw key text.

Acceptance:
- Unicode input is rune-safe and cursor state remains inside buffer bounds.

### `func (e *Editor) insert(text string)`

Logic:
- Insert text at rune cursor, splitting multiline paste into editor lines.

Acceptance:
- Unicode input is rune-safe and cursor state remains inside buffer bounds.

### `func (e *Editor) backspace()`

Logic:
- Delete previous rune or merge with previous line.

Acceptance:
- Unicode input is rune-safe and cursor state remains inside buffer bounds.

### `func (e *Editor) forwardDelete()`

Logic:
- Delete next rune or merge next line.

Acceptance:
- Unicode input is rune-safe and cursor state remains inside buffer bounds.

### `func (e *Editor) move(delta int)`

Logic:
- Move cursor horizontally and clamp to the current line rune length.

Acceptance:
- Unicode input is rune-safe and cursor state remains inside buffer bounds.

### `func clampRuneIndex(index int, length int) int`

Logic:
- Clamp a requested rune index to `[0, length]` before slicing line buffers.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (e *Editor) changed()`

Logic:
- Notify the optional change handler after a mutation that changes editor text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.
