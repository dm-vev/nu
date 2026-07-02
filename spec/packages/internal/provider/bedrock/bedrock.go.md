# `internal/provider/bedrock/bedrock.go`

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

Amazon Bedrock adapter.

## Code Style

AWS credential resolution is injected behind a small signer/client boundary so
tests avoid real AWS config.

## Functions

### `NewClient(cfg Config) *Client`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Support region, endpoint override, bearer token, profile/IAM signer, and HTTP.

Acceptance:

- supports region, endpoint override, bearer token, profile/IAM signer, and HTTP
  client.

### `(*Client) Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- Validate `provider.Request` with `ValidateRequest`.
- Resolve AWS/bearer credentials through injected signer/client boundary.
- Convert Nu messages to Bedrock Converse messages and system blocks.
- Convert Nu tools to Bedrock tool config.
- Insert cache points when model/settings require them.
- Sign and send streaming Converse request.
- Decode Bedrock stream events and map content block deltas, reasoning where
  available, tool use, metadata usage, stop reason, throttling, and errors to
  `provider.Event`.
- Classify AWS throttling/auth/quota/region/model errors for retry layer.

Acceptance:

- converts Bedrock converse stream events to neutral provider events;
- supports forced cache flags;
- emits only events allowed by `spec/protocols/provider-stream.md`.

Tests:

- `TestBedrockRequestShape`
