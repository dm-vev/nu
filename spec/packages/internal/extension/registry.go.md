# `internal/extension/registry.go`

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

Track extension-registered tools, commands, keybindings, flags, providers, and
renderers.

## Code Style

Registration validates names before mutating state.

## Functions

### `(*Registry) RegisterTool(extID string, def ToolDef) error`

Logic:

- Validate extension id, tool name, description, schema, and execution mode.
- Reject duplicate tool names across built-ins and extensions according to registry policy.
- Record ownership so tool execution and diagnostics can attribute the source extension.
- Store an immutable definition copy rather than caller-owned pointers.

Acceptance:

- rejects invalid or duplicate names;
- records extension ownership.

### `(*Registry) Snapshot() Snapshot`

Logic:

- Copy registered tools, commands, flags, keybindings, hooks, and UI contributions.
- Sort every collection deterministically by name/source for stable UI and tests.
- Return immutable slices/maps so app code cannot mutate the live registry.
- Include ownership metadata needed for diagnostics and conflict reporting.

Acceptance:

- returns immutable copies for app use.

Tests:

- `TestNUF160ExtensionRegistersTool`
