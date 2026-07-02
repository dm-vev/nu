# `internal/config/settings.go`

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

Load, merge, validate, and save settings.

## Code Style

Use explicit structs with `json.RawMessage` only for extension-owned fields.
Missing settings use zero values plus defaults in one place.

## Functions

### `LoadSettings(ctx context.Context, fs FS, paths Paths, trust TrustDecision) (Settings, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Load global settings.
- Load project settings only when trusted.
- Report malformed JSON with file path.
- Do not read real home in tests.

Acceptance:

- loads global settings;
- loads project settings only when trusted;
- reports malformed JSON with file path;
- does not read real home in tests.

### `MergeSettings(global, project Settings, overrides Overrides) Settings`

Logic:

- Start from built-in defaults and layer global settings, trusted project settings, then CLI overrides.
- Merge maps by key with later sources replacing earlier values.
- Treat slices/lists as replacing the previous value unless a field explicitly defines append semantics.
- Keep extension-owned raw JSON grouped by extension id so unknown settings survive load/save.

Acceptance:

- applies precedence: defaults < global < trusted project < CLI overrides.

### `SaveSettings(ctx context.Context, fs FS, path string, settings Settings) error`

Logic:

- Check context before filesystem work.
- Normalize settings into the persisted JSON struct and omit runtime-only fields.
- Create the parent directory through the injected filesystem.
- Write stable indented JSON with deterministic map ordering where Go encoding does not already guarantee it.

Acceptance:

- writes stable indented JSON;
- creates parent dir when needed.

Tests:

- `TestNUF010ProjectSettingsIgnoredWithoutTrust`
- `TestNUF010CLIOverridesSettings`
