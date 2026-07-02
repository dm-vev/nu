# `internal/config/paths.go`

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

Resolve global and project Nu paths.

## Code Style

Pure path functions where possible. Use `filepath`, not string concatenation.

## Functions

### `DefaultPaths(home, cwd string) Paths`

Logic:

- Clean caller-provided `home` and `cwd` without consulting process globals.
- Set global root to `$home/.nu/agent` and project root to `$cwd/.nu`.
- Derive auth, trust, settings, sessions, resources, packages, and extension paths from those roots.
- Keep path construction pure so tests can provide temp homes and relative cwd values.

Acceptance:

- resolves `~/.nu/agent`, `.nu`, auth, trust, settings, sessions, resources;
- handles relative cwd/home inputs by cleaning them.

### `ExpandPath(base, value string) (string, error)`

Logic:

- Reject an empty value when the caller marks the setting as required.
- Expand `~` against the configured home/base context, not the real user home.
- Return absolute paths unchanged after cleaning.
- Join relative paths to `base` and clean the result without checking existence.

Acceptance:

- supports absolute, relative, and `~` paths;
- rejects empty paths when caller marks them required.

Tests:

- `TestConfigDefaultPaths`
- `TestConfigExpandPath`
