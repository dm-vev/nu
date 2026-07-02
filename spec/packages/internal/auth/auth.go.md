# `internal/auth/auth.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Phase 3 auth resolves provider credentials without network calls.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Load `auth.json` credentials and resolve provider API keys for provider
adapters. Tests must never read the real home directory or real credentials.

## Code Style

Stdlib only. Keep command interpolation explicit and bounded by caller context.
Do not log or return credential values in errors.

## Types

### `type Credential`

Logic:

- Hold `api_key`, `api_key_env`, and `api_key_command` values from `auth.json`.
- Keep JSON names stable.

Acceptance:

- can represent literal, env-interpolated, and command-produced keys.

### `type Store`

Logic:

- Hold provider id to credential mapping.
- Store process environment as a map for deterministic tests.

Acceptance:

- does not consult process environment directly after construction.

## Functions

### `Load(path string, env []string) (Store, error)`

Logic:

- Return an empty store when the auth file is missing.
- Decode `{ "providers": { "<provider>": Credential } }`.
- Build an environment map from `KEY=value` entries.
- Wrap malformed JSON and read errors.

Acceptance:

- missing file is not an error;
- malformed files return a contextual error.

### `ResolveAPIKey(ctx context.Context, providerID string) (string, bool, error)`

Logic:

- Prefer auth-file credentials over environment fallback.
- Resolve literal `api_key` after `${VAR}` interpolation.
- Resolve `api_key_env` by environment name.
- Resolve `api_key_command` by running `sh -c` under caller context and
  trimming whitespace.
- Fall back to known provider env names when no auth-file credential exists.

Acceptance:

- auth file beats environment;
- env interpolation works;
- command interpolation works;
- errors do not include secret values.

Tests:

- `TestNUF020AuthFileBeatsEnvironment`
- `TestNUF020EnvInterpolation`
- `TestNUF020CommandInterpolation`

