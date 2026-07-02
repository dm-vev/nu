# `internal/provider/errors.go`

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

Classify provider errors for retry and user diagnostics.

## Code Style

Use `errors.Is`/`errors.As` compatible types. Do not parse display strings in
agent retry code.

## Functions

### `Classify(err error) ErrorClass`

Logic:

- Return `cancelled` for `context.Canceled`, context deadline, or adapter
  cancellation errors.
- Unwrap typed provider errors first.
- Map HTTP 401/403 to `auth`, unsupported-model style errors to `unsupported`,
  408/409/425/429/5xx to retryable classes, and quota-specific payload codes to
  `quota`.
- Treat malformed provider streams as `fatal` unless adapter marks them
  transient.
- Never inspect API keys, headers, or raw request body.

Acceptance:

- distinguishes transient, auth, quota, unsupported, rate-limit, cancellation,
  and fatal errors.

### `RetryAfter(err error) (time.Duration, bool)`

Logic:

- Unwrap typed provider errors and HTTP metadata.
- Prefer explicit `retry_after_ms` from normalized error event.
- Fall back to HTTP `Retry-After` seconds or HTTP-date when available.
- Return `false` for negative, zero, absent, or unparsable values.

Acceptance:

- extracts provider-requested retry delays when present.

Tests:

- `TestProviderErrorClassification`
