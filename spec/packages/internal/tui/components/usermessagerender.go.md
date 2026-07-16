# `internal/tui/components/usermessagerender.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Renders user messages as boxed Markdown with OSC markers.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by component tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Render user prompts in the Pi-like highlighted block.

## Constants

### `const userMessageOsc133ZoneStart`, `userMessageOsc133ZoneEnd`, `userMessageOsc133ZoneFinal`

Logic:
- Define Pi-compatible prompt-zone markers for user turns.

Acceptance:
- Markers wrap non-empty user output.

## Functions

### `func (m *UserMessage) Render(width int) []string`

Logic:
- Create a padded box with the configured background.
- Render message text through the Markdown component.
- Add OSC 133 start/end/final markers around the result.

Acceptance:
- Markdown output is width-bounded and grouped as one prompt zone.
