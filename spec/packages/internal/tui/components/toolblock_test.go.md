# `internal/tui/components/toolblock_test.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Tests command, failure, and patch rendering.

## TODO

- [x] Tests cover success, command failure, and patch color paths.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Lock in Pi-like tool block behavior.

## Tests

### `TestToolBlockToolBlockRendersBashCommandAndOutput`

Logic:
- Render a successful bash result.

Acceptance:
- Output includes `$ command`, command output, and success background.

### `TestToolBlockToolBlockRendersFailedCommandWithErrorBackground`

Logic:
- Render a bash result with non-zero exit code.

Acceptance:
- Output includes error background and error text color.

### `TestToolBlockToolBlockRendersPatchDiffColors`

Logic:
- Render an edit patch result.

Acceptance:
- Added and removed lines contain configured diff colors.
