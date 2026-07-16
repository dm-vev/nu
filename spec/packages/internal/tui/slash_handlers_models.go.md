# `internal/tui/tui_slash_handlers_models.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Scoped-model display logic was split from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Render visible model candidates and identify the current provider/model.

## Code Style

Keep this display-only and copy state under the app mutex before rendering.

## Owned Logic

- `modelSummary` contains only provider, ID, display label, and current marker.
- `summarizeModels` converts registry models without exposing full metadata.
- `scopedModelsCommandText` renders an empty-state message or Markdown table.

## Acceptance

- The active provider/model is marked exactly once when present.
- Empty available models produce a clear local response.

## Tests

- `TestTUISlashModelOpensSelectorAndSelectsModel`
- `TestTUISlashModelExactMatchSelectsWithoutMenu`
- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
