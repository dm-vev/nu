# `internal/provider/google/google.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Phase 3 Google adapter covers GenerateContent request shape and simple SSE parsing.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement Google Gemini GenerateContent streaming adapter without SDK dependency.

## Functions

### `BuildGenerateContentPayload(req provider.Request) (map[string]any, error)`

Logic:

- Convert user messages to `role: user`.
- Convert assistant messages to `role: model`.
- Convert text into `parts: [{ text: ... }]`.

Acceptance:

- matches GenerateContent request shape.

### `Stream(ctx context.Context, req provider.Request) (<-chan provider.Event, error)`

Logic:

- POST to `models/{model}:streamGenerateContent?alt=sse&key=...`.
- Parse candidate text parts as text deltas.
- Parse function calls as one complete tool call start/delta/end sequence.

Acceptance:

- request shape and simple stream parsing are test-covered.

Tests:

- `TestGoogleGenerateContentRequestShape`

