# `internal/provider/bedrock/bedrock.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Phase 3 Bedrock adapter covers ConverseStream request shape, toolUse/toolResult history, SigV4 request signing, event-stream frame size limits, and minimal event-stream JSON payload parsing.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the Bedrock ConverseStream request builder and SigV4 HTTP setup
without adding the AWS SDK.

## Functions

### `BuildConversePayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user/assistant messages into Bedrock `messages`.
- Convert text into `content: [{ text: ... }]`.
- Convert assistant tool-call history into `toolUse` content blocks.
- Convert tool results into user-role `toolResult` content blocks.

Acceptance:

- matches ConverseStream request shape;
- preserves Bedrock toolUse history before toolResult messages.

### `converseMessage(message provider.Message) map[string]any`

Logic:

- Convert assistant tool-call history into assistant-role `toolUse` content.
- Convert tool results into user-role `toolResult` content.
- Preserve simple text messages as Bedrock text content blocks.

Acceptance:

- Bedrock payload messages preserve toolUse/toolResult order.

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
- Reject eventstream frames above the package maximum before allocating the
  frame body.

Acceptance:

- request shape is test-covered;
- signing is deterministic;
- oversized remote frames are rejected before large allocation;
- mocked stream frames normalize text and done events.

### `readEventStreamPayload(r io.Reader) ([]byte, error)`

Logic:

- Read and validate the AWS eventstream prelude.
- Reject invalid lengths and prelude CRC mismatches.
- Reject frames larger than the package maximum before allocating the frame body.
- Validate message CRC and return only the JSON payload slice.

Acceptance:

- malformed frames fail with contextual errors;
- oversized remote frames cannot force large allocations.

### `decodeJSONOrText(raw string) any`

Logic:

- Decode JSON arguments/results when possible.
- Fall back to `{ "text": raw }` when the value is not valid JSON.

Acceptance:

- malformed tool arguments do not panic or fail payload construction.

Tests:

- `TestBedrockConverseRequestShape`
- `TestBedrockPayloadIncludesToolUseAndResult`
- `TestBedrockSignAddsAuthorization`
- `TestBedrockRejectsOversizedEventFrame`
