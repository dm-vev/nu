# `internal/provider/google/generate.go`

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

Google Generative AI adapter.

## Code Style

Keep Google role/content conversion isolated. Do not leak Google SDK types into
core packages.

## Functions

### `NewClient(cfg Config) *Client`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Configure base URL, key resolver, headers, and HTTP client.

Acceptance:

- configures base URL, key resolver, headers, and HTTP client.

### `(*Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Resolve API key and headers immediately before request execution.
- Convert system prompt and Nu messages to Gemini contents and roles.
- Convert image blocks to inline data parts.
- Convert Nu tools to function declarations.
- Map thinking level to Gemini thinking/config fields when supported.
- Create streaming generate-content HTTP request.
- Decode stream frames and map candidates, text parts, function calls, safety
  blocks, usage metadata, and errors to `provider.Event`.
- Treat blocked/safety finish reasons as terminal done/error according to
  provider classification.

Acceptance:

- converts text, image, thinking, tool calls, safety/error responses, and usage;
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestGoogleGenerateRequestShape`
