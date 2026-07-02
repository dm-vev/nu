# `internal/auth/auth.go`

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

Load, save, and query provider credentials.

## Code Style

Never log secrets. Keep file permission handling close to writes.

## Functions

### `LoadStore(ctx context.Context, fs FS, path string, env map[string]string) (*Store, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Auth file values override process environment.
- Missing auth file is not an error.
- Malformed auth file returns path-qualified error.

Acceptance:

- auth file values override process environment;
- missing auth file is not an error;
- malformed auth file returns path-qualified error.

### `(*Store) Credential(provider string) (Credential, bool)`

Logic:

- Normalize provider id using the same id form as model/provider registry.
- Return a provider-specific credential from the auth file when present.
- Fall back to provider environment variable resolution from the store snapshot.
- Return `false` without probing process environment when neither source exists.

Acceptance:

- resolves provider-specific auth file entry first, env second.

### `(*Store) Save(ctx context.Context) error`

Logic:

- Check context before writing.
- Merge updated credentials with unknown auth-file fields preserved from load.
- Write the auth file with `0600` permissions where the platform supports it.
- Never serialize resolved environment-only secrets back into the auth file unless explicitly added.

Acceptance:

- writes `0600` where supported;
- preserves unknown credential fields.

Tests:

- `TestNUF020AuthFileBeatsEnvironment`
