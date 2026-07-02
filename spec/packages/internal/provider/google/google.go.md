# `internal/provider/google/google.go`

## Status

Current: IMPLEMENTED
Implementation Commit: c64b048
Implementation Comments: Phase 3 Google adapter covers GenerateContent request shape, function-call history, function responses, and simple SSE parsing.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement Google Gemini GenerateContent streaming adapter without SDK dependency.

## Functions

### `BuildGenerateContentPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user messages to `role: user`.
- Convert assistant messages to `role: model`.
- Convert text into `parts: [{ text: ... }]`.
- Convert assistant tool-call history into `functionCall` parts.
- Convert tool results into `functionResponse` parts.

Acceptance:

- matches GenerateContent request shape;
- preserves function-call history before function responses.

### `generateContentMessage(message provider.Message) map[string]any`

Logic:

- Convert assistant tool-call history into model-role `functionCall` parts.
- Convert tool results into user-role `functionResponse` parts.
- Convert assistant text roles to `model` and preserve user text as `user`.

Acceptance:

- Gemini payloads preserve function-call/function-response order.

### `decodeJSONOrText(raw string) any`

Logic:

- Decode JSON arguments/results when possible.
- Fall back to `{ "text": raw }` when the value is not valid JSON.

Acceptance:

- malformed tool arguments do not panic or fail payload construction.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to `models/{model}:streamGenerateContent?alt=sse&key=...`.
- Parse candidate text parts as text deltas.
- Parse function calls as one complete tool call start/delta/end sequence.

Acceptance:

- request shape and simple stream parsing are test-covered.

Tests:

- `TestGoogleGenerateContentRequestShape`
- `TestGooglePayloadIncludesFunctionCallAndResponse`
