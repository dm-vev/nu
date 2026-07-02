# `internal/extension/ui.go`

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

Bridge extension UI requests to TUI/headless mode.

## Code Style

Core returns typed UI requests. TUI renders them; headless mode returns safe
defaults or errors.

## Functions

### `HandleUIRequest(ctx context.Context, req UIRequest, bridge Bridge) (UIResponse, error)`

Logic:

- Validate UI request type and required fields.
- If bridge mode is headless and request is interactive, return provided default
  or a typed unsupported-interactive error.
- For `notify`, `status`, and `widget`, forward and return immediately.
- For `select`, `confirm`, and `input`, call bridge and wait for user result or
  context cancellation.
- For `custom`, require interactive TUI bridge and pass component descriptor to
  TUI overlay/custom component manager.
- Return response frame payload to extension host.

Acceptance:

- supports select, confirm, input, notify, status, widget, and custom component
  requests;
- rejects interactive-only requests in print/JSON mode unless default exists.

Tests:

- `TestExtensionUIHeadlessRejectsPrompt`
