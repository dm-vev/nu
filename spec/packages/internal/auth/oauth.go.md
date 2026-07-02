# `internal/auth/oauth.go`

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

Represent OAuth credentials and refresh flow for subscription providers.

## Code Style

Provider-specific OAuth endpoints live in provider packages; this file owns
generic token storage and refresh orchestration.

## Functions

### `RefreshIfNeeded(ctx context.Context, cred OAuthCredential, client OAuthClient) (OAuthCredential, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Refreshes before expiry with a small clock skew.
- Return original credential when still valid.
- Wraps provider refresh errors.

Acceptance:

- refreshes before expiry with a small clock skew;
- returns original credential when still valid;
- wraps provider refresh errors.

Tests:

- `TestOAuthRefreshWhenExpired`
- `TestOAuthKeepsValidToken`
