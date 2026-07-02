# `internal/tool/tool.go`

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

Core tool interface and result types.

## Code Style

Keep interface minimal. Tool implementations own their argument structs.

## Types

### `Tool`

Logic:

- Define the minimal interface consumed by agent tool execution.
- Expose name, description, JSON schema, execution mode, ownership/source, and execute method.
- Keep UI labels and provider-specific schema conversion outside the interface.
- Require execute implementations to return `Result` instead of printing or panicking.

Acceptance:

- exposes name, description, schema, execution mode, and execute method.

### `Result`

Logic:

- Store content blocks for model-visible output.
- Store structured details for UI/export diagnostics and hidden metadata.
- Represent tool failure as data with an error flag/code/message, not as a Go error after execution starts.
- Preserve provider tool call id association in the agent layer, not in the generic result type.

Acceptance:

- contains content blocks, details, and error state.

## Functions

No registry or execution helpers belong in this file. Registry logic stays in
`registry.go`; validation stays in `schema.go`.

Tests:

- covered by registry and agent tool tests.
