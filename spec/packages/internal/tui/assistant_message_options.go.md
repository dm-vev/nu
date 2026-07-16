# `internal/tui/assistant_message_options.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Normalizes assistant text/thinking/tool style callbacks.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Nil style callbacks normalize safely.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Configure assistant text Markdown, thinking Markdown, and tool block styles.

## Acceptance Criteria

- Missing text styles fall back to identity output.
- Missing thinking styles fall back to text styles.
- Missing tool text/diff styles fall back to readable text output.

## Types

### `type AssistantMessageOptions struct`

Logic:
- Carry padding plus style hooks for text, headings, strong/emphasis/code,
  thinking, tool backgrounds, tool title/output/error, and diff lines.

Acceptance:
- App-level palette decisions stay outside the component package.

## Functions

### `func assistantMessageNormalizeOptions(opts AssistantMessageOptions) AssistantMessageOptions`

Logic:
- Clamp padding and fill nil callbacks with deterministic defaults.

Acceptance:
- Component rendering cannot panic on missing style callbacks.

### `func assistantMessageIdentity(value string) string`

Logic:
- Return text unchanged.

Acceptance:
- Used as the default style callback.
