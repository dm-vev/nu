# `internal/provider/sse.go`

## Status

Current: PLANNED
Implementation Commit: TBD
Implementation Comments: Shared SSE reader keeps HTTP adapters small.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Parse server-sent event streams used by OpenAI, Anthropic, Google, and compatible
providers.

## Code Style

Small stdlib parser. Preserve event order. Ignore comments and empty keepalive
frames.

## Types

### `type SSEvent`

Logic:

- Store optional event name and joined data payload.

Acceptance:

- supports `event:` and multi-line `data:` frames.

## Functions

### `ReadSSE(ctx context.Context, r io.Reader, emit func(SSEvent) error) error`

Logic:

- Read LF-delimited SSE frames.
- Strip one trailing `\r`.
- Join repeated `data:` lines with newline.
- Ignore comment lines and empty frames.
- Stop on context cancellation.

Acceptance:

- preserves event order;
- returns contextual scan and emit errors.

Tests:

- covered through provider adapter stream tests.

