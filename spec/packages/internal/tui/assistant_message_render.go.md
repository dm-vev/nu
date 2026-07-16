# `internal/tui/assistant_message_render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Composes assistant text, thinking, and tool components.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Text, thinking, and tool paths are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Render structured assistant parts in order.

## Code Style

Components render `[]string` at supplied width and must not write to the
terminal directly. Keep OSC marker behavior local to this file.

## Acceptance Criteria

- Text parts render through `components/markdown`.
- Thinking parts render through `components/thinking`.
- Tool parts render through `components/toolblock`.
- Assistant messages with tool parts do not receive outer OSC 133 prompt-zone markers.
- Assistant messages without tool parts do receive OSC 133 markers.

## Constants

### `const assistantMessageOsc133ZoneStart`, `assistantMessageOsc133ZoneEnd`, `assistantMessageOsc133ZoneFinal`

Logic:
- Define Pi-compatible prompt-zone markers for plain assistant output.

Acceptance:
- Markers are applied only when no tool parts exist.

## Functions

### `func (m *AssistantMessage) Render(width int) []string`

Logic:
- Add Pi-like leading spacing for visible text/thinking content.
- Render every text/thinking/tool part through its dedicated component.
- Insert spacing after thinking only when another text/thinking part follows.
- Let tool blocks provide their own preceding spacer.
- Add OSC 133 markers only when the message has no tool parts.

Acceptance:
- Mixed assistant messages keep part order and visible widths stay bounded.

### `func (m *AssistantMessage) hasTextOrThinkingContent() bool`

Logic:
- Return true when any text/thinking content is non-empty.

Acceptance:
- Tool-only assistant turns do not get an extra assistant spacer.

### `func (m *AssistantMessage) hasTextOrThinkingContentAfter(index int) bool`

Logic:
- Check if a later text/thinking part will render visible content.

Acceptance:
- Thinking-to-text spacing is added only when needed and does not duplicate tool spacing.

### `func (m *AssistantMessage) hasToolParts() bool`

Logic:
- Return true when any part is a tool block.

Acceptance:
- OSC marker policy can match Pi behavior.
