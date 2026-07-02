# `internal/provider/openai/sse.go`

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

Shared SSE decoder for OpenAI-compatible streams.

## Code Style

Parser is line-based and tested with chunk boundaries. No provider business
logic here.

## Functions

### `DecodeSSE(r io.Reader) (<-chan SSEEvent, <-chan error)`

Logic:

- Decode input in one pass into typed data.
- Preserve enough location/context for diagnostics.
- Return typed errors instead of exiting or printing.
- Handle split lines and multi-line data.
- Ignores comments.
- Emits done marker.
- Return malformed frame errors.

Acceptance:

- handles split lines and multi-line data;
- ignores comments;
- emits done marker;
- returns malformed frame errors.

Tests:

- `TestOpenAISSEDecodeChunkedFrames`
