# `internal/app/cli/help.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Stable help/version rendering lives with app mode dispatch.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render help and version output.

## Code Style

Pure formatting functions. No direct stdout writes.

## Functions

### `Help(extensionFlags []ExtensionFlag) string`

Logic:

- Render built-in modes and flag groups in a stable order.
- Include model, session, tool, resource, package, config, update, export,
  share, and RPC options.
- Append extension flags after built-ins with source labels.
- Return a string only; writing to stdout belongs to the app layer.

Acceptance:

- includes modes, model options, session options, tool options, resource options,
  package commands, share command, and RPC options;
- includes extension flags when present.

### `Version(info VersionInfo) string`

Logic:

- Format binary name, version, commit, build date, and Go runtime data into one line.
- Omit empty optional fields without changing the order of present fields.
- Avoid color, localization, and terminal width logic so scripts can parse the output.
- Return a trailing-newline-free string; caller decides printing.

Acceptance:

- stable one-line output suitable for scripts.

Tests:

- `TestCLIHelpMentionsCoreModes`
- `TestCLIHelpIncludesExtensionFlags`
