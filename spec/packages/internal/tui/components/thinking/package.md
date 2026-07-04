# `internal/tui/components/thinking`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Render model reasoning as Markdown with a distinct thinking style. The package
exists so assistant message rendering can keep text and reasoning as separate
content parts.

## Files

- `options.go`: padding and thinking style callbacks.
- `thinking.go`: component state.
- `render.go`: Markdown delegation with thinking styles.
- `thinking_test.go`: gray/italic style coverage.

## Acceptance Criteria

- Thinking text supports Markdown.
- Thinking output uses the configured dim/italic style.
- `go test ./internal/tui/components/thinking` passes.
