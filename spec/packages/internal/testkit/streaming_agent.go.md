# `internal/testkit/streaming_agent.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -

## Purpose

Provide a deterministic fake `contracts.StreamingAgent` for Nu app/RPC/TUI
tests. It records prompts and emits scripted SDK agent events without recreating
provider requests or a model/tool loop.

## Tests

Used by app and RPC mode tests.
