# `internal/agentui/events.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -

Implementation Comments: Maps SDK AgentStreamEvent into existing Nu TUI/RPC events.

## Purpose

Translate SDK agent content/thinking/tool/result/error/complete events into the
existing Nu TUI/RPC event shape without performing backend work.

## Acceptance

- visible content accumulates into `turn_end`;
- thinking and tool state remain structured;
- SDK errors do not produce successful `turn_end`;
- empty SDK boundary events do not create visible text.
