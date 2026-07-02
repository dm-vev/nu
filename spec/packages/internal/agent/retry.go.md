# `internal/agent/retry.go`

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

Agent-level retry policy for transient provider failures.

## Code Style

Inject sleeper/clock for tests. Do not sleep directly in tests.

## Functions

### `RunWithRetry(ctx context.Context, policy RetryPolicy, op func(context.Context) error) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Initialize local state, then enter the smallest required loop.
- Stop on context cancellation, terminal command, or unrecoverable error and clean up owned resources.
- Retry only transient provider errors.
- Emits retry start/end events through callback.
- Honor max attempts and base delay.
- Refuse provider retry-after above cap.

Acceptance:

- retries only transient provider errors;
- emits retry start/end events through callback;
- honors max attempts and base delay;
- refuses provider retry-after above cap.

### `NextDelay(policy RetryPolicy, attempt int, err error) (time.Duration, bool, error)`

Logic:

- Classify the provider error before calculating a delay.
- Return `false` when the error is non-transient, attempts are exhausted, or context cancellation caused the failure.
- Prefer provider `retry-after` only when it is positive and below policy cap.
- Otherwise compute deterministic exponential backoff from base delay and attempt number.

Acceptance:

- deterministic exponential backoff unless provider retry-after is allowed.

Tests:

- `TestNUF053RetriesTransientError`
- `TestNUF053LongRetryAfterFailsFast`
