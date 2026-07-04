# `internal/tui/ansi/style.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines SGR style constants used by TUI components.

## TODO

- [x] File exists in the split `internal/tui/ansi` architecture.
- [x] Constants are covered through component render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Provide reusable ANSI SGR constants for foregrounds, backgrounds, bold, italic,
inverse, and reset sequences.

## Acceptance Criteria

- Foreground styles restore with `DefaultFG`.
- Background styles restore with `DefaultBG`.
- Added constants support Markdown, thinking, and tool block rendering.

## Constants

### `const Reset`, `Bold`, `BoldOff`, `Italic`, `ItalicOff`, `Inverse`, `InverseOff`

Logic:
- Define common SGR style toggles.

Acceptance:
- Markdown and editor rendering can compose styles without hard-coded sequences.

### `const DefaultFG`, `DefaultBG`

Logic:
- Restore terminal foreground or background without resetting all SGR state.

Acceptance:
- Background boxes remain filled while inner text changes foreground.

### `const Text`, `Muted`, `Dim`, `Green`, `Red`, `Yellow`, `Context`

Logic:
- Define Nu's foreground palette.

Acceptance:
- App style callbacks do not repeat raw escape sequences.

### `const UserMessageBG`, `ToolPendingBG`, `ToolSuccessBG`, `ToolErrorBG`

Logic:
- Define boxed message/tool backgrounds.

Acceptance:
- Tool and user message components can render Pi-like filled blocks.
