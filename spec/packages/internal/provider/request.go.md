# `internal/provider/request.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Basic provider request validation exists; tools, images, thinking, and cache hints are pending.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Provider-neutral request/response structs.
These structs are the only input provider adapters receive from the agent.

## Code Style

Keep close to Nu message model, not any one vendor payload.

## Types

### `Request`, `Tool`, `ToolSchema`, `Options`

Logic:

- Represent model, messages, tools, system prompt, max tokens, thinking,
  transport, cache hints, and provider metadata exactly once.
- Keep vendor-specific payload names out of the shared request.
- Carry capability decisions already resolved by `internal/model`.

Acceptance:

- contains model, messages, tools, system prompt, max tokens, thinking,
  transport, cache hints, and provider metadata.

## Functions

### `ValidateRequest(req Request) error`

Logic:

- Require provider, API, model, and at least one message.
- Verify every tool has name, description, and object schema.
- Verify image content is present only when model capabilities include image
  input.
- Verify thinking options are already mapped by `internal/model`.
- Return a typed validation error that includes field path, not raw secret data.

Acceptance:

- catches missing model/provider before adapter HTTP code;
- catches unsupported image input before network request;
- never validates vendor-specific payload details.

Tests:

- provider adapter golden tests verify conversion from these structs.
