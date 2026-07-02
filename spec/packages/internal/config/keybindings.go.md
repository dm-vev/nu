# `internal/config/keybindings.go`

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

Load and normalize keybinding config.

## Code Style

Keep parser deterministic. Key strings normalize to lowercase modifier order.

## Functions

### `LoadKeybindings(ctx context.Context, fs FS, path string) (Keymap, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Merge user config over defaults.
- Migrate old action IDs.
- Report invalid keys without losing valid bindings.

Acceptance:

- merges user config over defaults;
- migrates old action IDs;
- reports invalid keys without losing valid bindings.

### `NormalizeKey(input string) (Key, error)`

Logic:

- Trim surrounding whitespace and split modifiers from the final key token.
- Normalize modifier aliases into the documented canonical order.
- Accept printable keys, named control keys, function keys, arrows, and terminal paste markers listed in config spec.
- Reject unknown modifiers, duplicate modifiers, empty keys, and ambiguous casing with a field-qualified error.

Acceptance:

- accepts documented modifiers and keys;
- rejects unknown modifiers and empty keys.

Tests:

- `TestNUF102KeybindingConfigOverridesDefault`
- `TestNUF102OldKeybindingIDsMigrate`
