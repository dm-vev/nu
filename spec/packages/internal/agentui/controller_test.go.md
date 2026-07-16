# `internal/agentui/controller_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -

## Purpose

Prove SDK stream translation and cancellation without importing provider
implementations or exercising a second backend loop.

## Tests

- `TestSDKStreamMapsContentThinkingAndTools`
- `TestAbortCancelsSDKRunner`
