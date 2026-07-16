# `internal/agentui/controller.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -

Implementation Comments: SDK StreamingAgent lifecycle adapter; no model/tool loop.

## Purpose

Adapt the Nu TUI/RPC lifecycle to the vendored SDK `contracts.StreamingAgent`.
This package owns busy/cancel/model-swap UI state only; it must not implement an
LLM or tool loop.

## Acceptance

- prompts delegate exactly once to SDK `RunStream`;
- abort cancels the active context;
- model changes rebuild the SDK agent while preserving memory;
- reset clears the scoped SDK memory;
- no provider-specific imports.
