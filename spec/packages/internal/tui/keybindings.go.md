# `internal/tui/keybindings.go`

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

Resolve input events to action IDs.

## Code Style

No hardcoded action checks in components. Components ask keymap for action.

## Functions

### `ActionFor(keymap Keymap, event InputEvent) (ActionID, bool)`

Logic:

- Normalize the input event into the same key representation used by config loading.
- Check user overrides before default bindings.
- Allow multiple keys to map to the same action but only one action per normalized key.
- Return `false` for text input events that are not bound commands.

Acceptance:

- supports multiple bindings per action;
- respects user overrides loaded by config.

Tests:

- `TestTUIActionForKeybinding`
