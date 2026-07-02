# `internal/tool/schema.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Minimal JSON schema support for tool arguments.

## Code Style

Use stdlib JSON. Do not implement a full JSON Schema validator unless tests
prove it is needed.

## Functions

### `Validate(schema Schema, raw json.RawMessage, dst any) error`

Logic:

- Require the raw arguments to decode as a JSON object.
- Check required fields before writing into `dst` so partial structs do not escape on failure.
- Reject unknown fields when strict mode is enabled.
- Decode into `dst` and return field-path errors that do not include secret-like values.

Acceptance:

- decodes object args into typed structs;
- rejects missing required fields;
- rejects unknown fields when schema requests strict mode.

Tests:

- `TestNUF051InvalidArgsReturnToolError`
