# `internal/tui/tui_messages.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Rebuilds chat from structured message parts.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Structured text/thinking/tool paths are covered by TUI tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Translate TUI message state into user and assistant components.

## Code Style

Keep app-state mutation here small and explicit. Rendering decisions belong in
the shared `tui/components` package after migration.

## Acceptance Criteria

- User prompts render as boxed Markdown messages.
- User prompt text uses the normal white text style, not the green accent style.
- Assistant messages render structured text/thinking/tool parts.
- Final text replacement does not duplicate text around tool blocks.

## Functions

### `func (a *App) rebuildChatLocked()`

Logic:
- Clear the chat container.
- For each user message, create a spacer and `usermessage` Markdown block.
- For each assistant message, create an `assistantmessage` structured renderer
  with app palette callbacks.
- Refresh the footer's display-only context estimate from all current message parts.

Acceptance:
- `TestTUIAppRendersStructuredMessageParts` covers mixed assistant content.
- `TestTUIAppRendersUserMessageTextWhite` covers the user text color.

### `func (a *App) appendAssistantDeltaLocked(delta string)`

Logic:
- Ignore empty deltas.
- Append text to the last assistant message or create an assistant message.

Acceptance:
- Streaming text remains ordered with other assistant parts.

### `func (a *App) appendAssistantThinkingLocked(delta string)`

Logic:
- Ignore empty deltas.
- Append thinking to the last assistant message or create an assistant message.

Acceptance:
- Thinking deltas render as gray/italic parts instead of normal text.

### `func (a *App) replaceLastAssistantLocked(value string)`

Logic:
- Replace final text only when the latest assistant message has no tool part.
- Append a new assistant text message when no assistant message exists.

Acceptance:
- Aggregated `turn_end.text` cannot duplicate text around tool blocks.

### `func (a *App) appendToolLocked(id string, name string, arguments string)`

Logic:
- Ensure the latest message is assistant-owned, then append a pending tool part.

Acceptance:
- `tool_start` immediately creates a visible tool block.

### `func (a *App) finishToolLocked(id string, result string, failed bool)`

Logic:
- Finalize the latest assistant tool part matching the id.

Acceptance:
- `tool_end` updates the visible block with output and success/error state.

### `func firstText(value Message) string`

Logic:
- Return the first text part from a message.

Acceptance:
- User message rendering gets prompt text without inspecting other part kinds.

### `func hasToolPart(value Message) bool`

Logic:
- Report whether a message contains at least one tool part.

Acceptance:
- Final text replacement can avoid tool-related duplication.

### `func estimateContextTokens(messages []Message) int`

Logic:
- Sum visible text, tool arguments, and tool results in runes.
- Return zero for empty history; otherwise return a conservative display-only estimate of one token per four runes, rounded up.

Acceptance:
- Footer usage changes when message content is added; provider-backed exact usage can replace this estimate later without changing footer rendering.
