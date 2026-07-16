# `internal/tui/markdown_block.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Parses Markdown block forms used by chat output.

## TODO

- [x] File exists in the split component architecture.
- [x] Headings, lists, quotes, and fenced code are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Handle line-oriented Markdown before wrapping.

## Functions

### `func markdownRenderBlocks(source string, opts MarkdownOptions) []string`

Logic:
- Normalize line endings, handle fenced code state, and style each block line.
- Recognize simple pipe tables as header, separator, and body rows.

Acceptance:
- Code fence contents are not interpreted as inline Markdown.
- Pipe table separator rows are structural and do not render as `---` text.

### `func markdownParseHeading(line string) (string, bool)`

Logic:
- Recognize ATX headings from `#` through `######`.

Acceptance:
- Only headings with a following space are accepted.

### `func markdownParseQuote(line string) (string, bool)`

Logic:
- Recognize `>` block quote prefixes and return trimmed quote text.

Acceptance:
- Quote lines render with a visible quote marker.

### `func markdownParseList(line string) (string, string, bool)`

Logic:
- Recognize unordered and ordered list markers.

Acceptance:
- Ordered list markers are preserved instead of rewritten.

### `func markdownParseTable(rawLines []string, start int, opts MarkdownOptions) ([]string, int, bool)`

Logic:
- Recognize a pipe table only when a row is followed by a Markdown separator row.
- Collect following non-blank pipe rows.
- Render rows with padded visible columns and inline Markdown inside cells.

Acceptance:
- `TestMarkdownMarkdownRendersPipeTables` fails if tables render as raw paragraph text.

### `func markdownSplitTableRow(line string) []string`

Logic:
- Trim outer pipes and split one table row into trimmed cells.

Acceptance:
- Leading and trailing table pipes are optional.

### `func markdownIsTableSeparator(line string) bool`

Logic:
- Accept separator cells made only from `-`, `:`, and spaces.

Acceptance:
- Header separator rows are not rendered as content.

### `func markdownTableWidths(rows [][]string, opts MarkdownOptions) []int`

Logic:
- Compute each column width from the visible width of inline-rendered cells.

Acceptance:
- Markdown markers inside cells do not widen columns.

### `func markdownRenderTableRow(row []string, widths []int, opts MarkdownOptions) string`

Logic:
- Render inline Markdown for each cell, right-pad to the column width, and join columns with two spaces.

Acceptance:
- Table rows remain terminal-width-safe after wrapping.
