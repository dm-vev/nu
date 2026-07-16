# `internal/tui/components/modelmenufilter.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Filtering and ordering logic.

## Purpose

Keep model search and current-first sorting out of render/input code.

## Functions

### `(*ModelMenu) refresh()`

Logic:
- Rebuild filtered candidates from query.
- Sort current model first, then by provider and id.
- Clamp selected index to the filtered range.

Acceptance:
- `TestModelMenuModelMenuRendersCurrentModelFirst` fails if current model is not first.

### `modelMenuMatchesQuery(entry model.Model, query string) bool`

Logic:
- Match all lowercase search tokens against provider, id, provider/id, aliases, and display name.

Acceptance:
- `TestModelMenuModelMenuFiltersByDisplayNameAndSelects` fails if display-name search stops working.

### `modelMenuModelSearchText(entry model.Model) string`

Logic:
- Build one searchable text string from model metadata.

Acceptance:
- Filtering does not duplicate metadata concatenation.

### `modelMenuModelDisplayName(entry model.Model) string`

Logic:
- Prefer display name, fallback to raw id.

Acceptance:
- Selector detail row is useful for custom model names.
