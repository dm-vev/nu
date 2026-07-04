# `internal/tui/components/modelmenu/filter.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Filtering and ordering logic.

## Purpose

Keep model search and current-first sorting out of render/input code.

## Functions

### `(*Menu) refresh()`

Logic:
- Rebuild filtered candidates from query.
- Sort current model first, then by provider and id.
- Clamp selected index to the filtered range.

Acceptance:
- `TestModelMenuRendersCurrentModelFirst` fails if current model is not first.

### `matchesQuery(entry model.Model, query string) bool`

Logic:
- Match all lowercase search tokens against provider, id, provider/id, aliases, and display name.

Acceptance:
- `TestModelMenuFiltersByDisplayNameAndSelects` fails if display-name search stops working.

### `modelSearchText(entry model.Model) string`

Logic:
- Build one searchable text string from model metadata.

Acceptance:
- Filtering does not duplicate metadata concatenation.

### `modelDisplayName(entry model.Model) string`

Logic:
- Prefer display name, fallback to raw id.

Acceptance:
- Selector detail row is useful for custom model names.
