# `internal/message/content.go`

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

Define content blocks shared by providers, sessions, tools, and UI.

## Code Style

Use tagged JSON structs. Keep custom unmarshalling minimal and tested.

## Types

### `Text`, `Image`, `Thinking`, `ToolCall`

Logic:

- Define a stable tagged JSON shape for each content block type.
- Keep image MIME type, size metadata, and base64 payload in separate fields; do not decode image bytes here.
- Store tool-call arguments as raw JSON object bytes so providers, sessions, and tools preserve exact values.
- Reject impossible local states in constructors or validation helpers, not during display/rendering.

Acceptance:

- JSON round-trips preserve type and fields;
- image MIME type and base64 data stay separate;
- tool call arguments preserve arbitrary JSON object values.

## Functions

No rendering or provider conversion functions belong in this file. If custom
JSON helpers are needed, add them here with round-trip tests first.

Tests:

- `TestNUF060MessageJSONRoundTrip`
- `TestNUF060ImageContentRoundTrip`
