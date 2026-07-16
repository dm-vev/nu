# `internal/tui/ansi/style.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines SGR style constants used by TUI components.

## TODO

- [x] File exists in the temporary flat implementation; target migration is `IN_PROGRESS`.
- [x] Constants are covered through component render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Provide reusable ANSI SGR constants for foregrounds, backgrounds, bold, italic,
inverse, and reset sequences.

## Acceptance Criteria

- Foreground styles restore with `ANSIDefaultFG`.
- Background styles restore with `ANSIDefaultBG`.
- Added constants support Markdown, thinking, and tool block rendering.

## Constants

### `const ANSIReset`, `ANSIBold`, `ANSIBoldOff`, `ANSIItalic`, `ANSIItalicOff`, `ANSIInverse`, `ANSIInverseOff`

Logic:
- Define common SGR style toggles.

Acceptance:
- Markdown and editor rendering can compose styles without hard-coded sequences.

### `const ANSIDefaultFG`, `ANSIDefaultBG`

Logic:
- Restore terminal foreground or background without resetting all SGR state.

Acceptance:
- Background boxes remain filled while inner text changes foreground.

### `const ANSIText`, `ANSIMuted`, `ANSIDim`, `ANSIGreen`, `ANSIRed`, `ANSIYellow`, `ANSIContext`

Logic:
- Define Nu's foreground palette.

Acceptance:
- App style callbacks do not repeat raw escape sequences.

### `const ANSIUserMessageBG`, `ANSIToolPendingBG`, `ANSIToolSuccessBG`, `ANSIToolErrorBG`

Logic:
- Define boxed message/tool backgrounds.

Acceptance:
- Tool and user message components can render Pi-like filled blocks.
