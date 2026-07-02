# `internal/message/message.go`

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

Define user, assistant, tool result, bash execution, custom, branch summary, and
compaction summary messages.

## Code Style

Keep message structs dumb. Derived display text belongs in TUI/export packages.

## Types

### `User`, `Assistant`, `ToolResult`, `BashExecution`, `Custom`, `BranchSummary`, `CompactionSummary`

Logic:

- Define message structs as persisted/domain data only; no provider conversion or display methods.
- Require role, timestamp, and stable id where the session layer persists the message.
- Store assistant provider, model, API kind, usage, stop reason, and thinking/tool-call metadata together.
- Store tool result call id, tool name, error state, content blocks, and structured details without flattening.

Acceptance:

- each message has role and timestamp;
- assistant includes provider/model/api/usage/stop reason;
- tool result includes tool call id, tool name, error state, content, details.

## Functions

No display, provider conversion, or storage functions belong in this file.
Message conversion helpers require separate specs in the consuming package.

Tests:

- `TestNUF060MessageJSONRoundTrip`
