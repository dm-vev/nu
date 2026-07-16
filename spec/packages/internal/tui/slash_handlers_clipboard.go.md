# `internal/tui/tui_slash_handlers_clipboard.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Clipboard selection and native command fallback were split from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Copy the latest assistant message through an available native clipboard command.

## Code Style

Use stdlib process lookup/execution and report unsupported clipboard environments locally.

## Owned Logic

- `handleCopySlash` handles empty history, copy failure, and success without reaching the model.
- `lastAssistantText` scans a locked message history and uses shared plain-text extraction.
- `copyToClipboard` tries `wl-copy`, `xclip`, `xsel`, then `pbcopy`.

## Acceptance

- User-only history is not copied.
- Clipboard text is passed on stdin and missing commands produce a visible local error.

## Tests

- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
- `TestTUISlashSessionDoesNotCallAgent`
