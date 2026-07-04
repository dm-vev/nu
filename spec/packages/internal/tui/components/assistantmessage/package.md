# `internal/tui/components/assistantmessage`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Assistant turn renderer for structured message parts. It renders assistant text
as Markdown, model thinking as gray/italic Markdown, and tool executions as
separate tool blocks.

## Files

- `options.go`: message, thinking, tool, and diff style callbacks.
- `message.go`: structured assistant message state.
- `render.go`: part-to-component composition and OSC marker policy.
- `message_test.go`: text, marker, thinking, and tool block checks.

## Acceptance Criteria

- Empty assistant turns render no lines.
- Assistant text supports Markdown.
- Thinking parts render separately from visible text.
- Tool parts render as separate blocks and suppress prompt-zone wrapping.
- Non-tool assistant turns include OSC 133 start/end/final markers.
