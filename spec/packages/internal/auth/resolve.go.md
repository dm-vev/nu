# `internal/auth/resolve.go`

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

Resolve configured secret values from literals, env interpolation, and shell
commands.

## Code Style

Keep command execution injected. Cache command-backed auth for process lifetime
only where the caller asks.

## Functions

### `ResolveValue(ctx context.Context, value string, env map[string]string, runner Runner) (string, bool, error)`

Logic:

- Handle literal values before expansion syntax.
- Resolve `$VAR` and `${VAR}` from the provided env map only.
- Interpret `$$` as a literal dollar and `$!` or `!command` as command-backed secret lookup.
- Trim one trailing newline from command stdout and suppress secret-like output on command errors.

Acceptance:

- supports `$VAR`, `${VAR}`, `$$`, `$!`, and `!command`;
- missing env returns `ok=false`;
- command stdout is trimmed of one trailing newline;
- command failure returns an error without exposing secret-like output.

Tests:

- `TestNUF020EnvInterpolation`
- `TestNUF020CommandInterpolation`
