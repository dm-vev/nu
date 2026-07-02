# `internal/provider/compat/openai.go`

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

OpenAI-compatible custom provider adapter.

## Code Style

Reuse OpenAI Chat/SSE helpers but keep compatibility flags explicit.

## Functions

### `NewOpenAICompatible(cfg Config) *Client`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Store custom base URL, auth header policy, custom headers, model-level API
  override, and compatibility flags without validating secrets or opening a
  network connection.

Acceptance:

- supports custom base URL, auth header toggle, custom headers, model-level API
  override, and compat flags.

### `(*Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Apply provider/model compatibility flags before payload conversion.
- Use OpenAI Chat conversion and SSE parser for compatible endpoints.
- Omit developer role, reasoning effort, strict schema, image fields, or auth
  header when compat flags say unsupported.
- Add configured custom headers after secret resolution.
- Convert SSE and non-2xx errors to normalized provider events/errors.

Acceptance:

- omits unsupported developer role or reasoning fields when configured;
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestCompatOpenAIOmitsUnsupportedFields`
