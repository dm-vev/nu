# `internal/provider/bedrock/bedrock.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Phase 3 Bedrock adapter covers ConverseStream request shape and SigV4 request signing.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the Bedrock ConverseStream request builder and SigV4 HTTP setup
without adding the AWS SDK.

## Functions

### `BuildConversePayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user/assistant messages into Bedrock `messages`.
- Convert text into `content: [{ text: ... }]`.

Acceptance:

- matches ConverseStream request shape.

### `Sign(req *http.Request, body []byte, creds Credentials, now time.Time) error`

Logic:

- Apply AWS Signature Version 4 headers for Bedrock runtime.
- Include session token when present.

Acceptance:

- deterministic signing is test-covered with fixed time.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to `/model/{modelId}/converse-stream`.
- Sign the request.
- Parse JSON event payloads from Bedrock event stream frames into provider
  events.

Acceptance:

- request shape is test-covered;
- signing is deterministic;
- mocked stream frames normalize text and done events.

Tests:

- `TestBedrockConverseRequestShape`
- `TestBedrockSignAddsAuthorization`

