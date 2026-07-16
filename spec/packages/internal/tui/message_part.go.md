# `internal/tui/message_part.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines text/thinking/tool parts and tool states.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Part kinds cover user text, assistant text, thinking, and tool blocks.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Describe one renderable message unit without importing component packages.

## Code Style

Keep this package data-only. Do not add rendering behavior here.

## Acceptance Criteria

- `PartText`, `PartThinking`, and `PartTool` are distinct.
- Tool parts carry id, name, raw arguments, raw result, and current tool state.

## Types

### `type PartKind string`

Logic:
- Identify how a message part must be rendered.

Acceptance:
- Switch statements in component packages use this type instead of raw strings.

### `type ToolState string`

Logic:
- Represent pending, success, and error execution states.

Acceptance:
- Tool block rendering can choose pending/success/error backgrounds from it.

### `type Part struct`

Logic:
- Store common text plus tool-specific metadata for one ordered content block.

Acceptance:
- The TUI can render command/patch blocks without parsing assistant prose.
