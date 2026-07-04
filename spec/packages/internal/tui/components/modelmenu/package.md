# `internal/tui/components/modelmenu`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-style model selector for `/model`.

## TODO

- [x] Package exists in the split `internal/tui` architecture.
- [x] Implementation is covered by package-level tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Render and drive the interactive model selection menu opened by `/model`.

## Code Style

Keep selector state local to this package. It may import `internal/model`, but it must not call providers, read auth files, or mutate the agent directly.

## Acceptance Criteria

- `/model` renders a selectable list of visible models.
- Search matches provider, id, aliases, provider/id, and display name.
- Arrow keys wrap, Enter selects, and Escape/Ctrl+C cancels.
