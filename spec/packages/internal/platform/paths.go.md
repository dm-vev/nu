# `internal/platform/paths.go`

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

Platform-specific path helpers.

## Code Style

Small functions wrapping OS conventions. Keep business path precedence in
`internal/config`.

## Functions

### `HomeDir() (string, error)`

Logic:

- Ask the OS/user package for the current user home directory once.
- Reject an empty home path with a clear platform error.
- Clean the returned path before handing it to config code.
- Do not create directories or apply Nu-specific defaults here.

Acceptance:

- returns OS user home or clear error.

### `ConfigDir(app string) (string, error)`

Logic:

- Resolve the platform config directory only for callers that explicitly request platform conventions.
- Validate `app` is non-empty and path-safe.
- Join and clean the config directory with the app name.
- Do not use this helper for the default `~/.nu` tree.

Acceptance:

- uses platform conventions only when caller asks, not for default `~/.nu`.

Tests:

- `TestPlatformPaths`
