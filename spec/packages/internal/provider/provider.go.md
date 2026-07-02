# `internal/provider/provider.go`

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

Define the provider adapter contract consumed by `internal/agent`.

## Code Style

Small consumer-owned interface. Provider implementations return concrete
clients.

## Types

### `type Streamer interface`

Logic:

- Expose only `Stream(ctx, req)` so adapters stay replaceable in tests.
- Require adapters to honor context cancellation and close their event channel.
- Return setup errors before starting a stream when request/auth configuration is invalid.
- Emit provider-neutral events defined in `stream.go` after streaming begins.

Acceptance:

- has `Stream(ctx context.Context, req Request) (<-chan Event, error)`;
- respects context cancellation.

### `type Registry`

Logic:

- Index provider streamers by provider id and API kind.
- Resolve the adapter selected by model metadata without inspecting CLI flags.
- Return typed unsupported-provider errors for missing adapter/API combinations.
- Keep registry immutable after construction so concurrent agent turns can share it.

Acceptance:

- resolves adapter by model API/provider.

## Functions

No provider implementation functions belong in this file. It owns only the
consumer-facing interface and minimal registry type.

Tests:

- covered by agent and provider contract tests.
