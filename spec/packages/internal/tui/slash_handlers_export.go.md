# `internal/tui/tui_slash_handlers_export.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Export/share/import parsing and renderers were split from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Snapshot and export current TUI messages, create share artifacts, and parse JSONL imports.

## Code Style

Use stdlib only, clone messages under lock, escape HTML, and write exports mode `0600`.

## Owned Logic

- `slashExportRecord` defines the minimal role/text interchange record.
- Export/share handlers resolve paths and select JSONL, Markdown, or HTML output.
- Snapshot/export helpers clone current state and create parent directories.
- Import/render helpers validate roles, bound scanner lines, escape HTML, and derive message plain text.

## Acceptance

- JSONL exports can be imported by the same code.
- Markdown/HTML exports are standalone and HTML content is escaped.
- Export errors identify the operation/path and do not call providers.

## Tests

- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
- `TestTUISlashSessionDoesNotCallAgent`
