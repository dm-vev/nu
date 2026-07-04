# `internal/tui/components/modelmenu/render.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Pi-style selector rendering.

## Purpose

Render model selector borders, hint/search rows, visible model rows, scroll info, and selected model detail.

## Functions

### `(*Menu) Render(width int) []string`

Logic:
- Return nil while hidden.
- Render bordered selector rows when visible.

Acceptance:
- Hidden selector contributes no layout rows.

### `(*Menu) renderItems(width int) []string`

Logic:
- Render empty state, visible model rows, scroll indicator, and selected model detail.

Acceptance:
- Long lists stay bounded by `MaxVisible`.

### `(*Menu) renderItem(entry model.Model, selected bool) string`

Logic:
- Render selected prefix, model id, provider badge, and current marker.

Acceptance:
- Current model is visibly marked.

### `(*Menu) visibleRange() (int, int)`

Logic:
- Center selected row within `MaxVisible` where possible.

Acceptance:
- Large lists scroll without changing total selector height unpredictably.

### `(*Menu) border(width int) string`

Logic:
- Render a full-width styled border.

Acceptance:
- Selector visually separates from chat and editor.

### `(*Menu) line(width int, text string) string`

Logic:
- Truncate and pad text to terminal width.

Acceptance:
- Selector rows never resize the frame horizontally.
