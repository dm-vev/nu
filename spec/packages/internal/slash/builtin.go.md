# `internal/slash/builtin.go`

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

Built-in slash command handlers.

## Code Style

Handlers call services; they do not contain storage/provider/TUI internals.

## Functions

### `RegisterBuiltins(reg *Registry, services Services)`

Logic:

- Register each `NUF-110` command name exactly once with its handler and help metadata.
- Bind handlers to injected services only; do not open files, terminals, or providers during registration.
- Fail fast on duplicate built-in names so extension collisions are handled by registry policy.
- Keep aliases explicit and tested rather than deriving names from function identifiers.

Acceptance:

- registers all commands listed in `NUF-110`;
- command names are stable and documented.

Tests:

- `TestSlashBuiltinsRegistered`
