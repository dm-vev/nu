# `internal/tui/style.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Defines Nu's dark green/black/gray/white TUI palette.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Palette functions are exercised through TUI render tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Centralize app-level style callbacks for text, Markdown, thinking, user blocks,
and tool blocks.

## Acceptance Criteria

- Main palette stays dark green, black, gray, and white.
- Thinking style is gray and italic.
- Tool success/pending/error backgrounds are separate callbacks.

## Functions

### `func green(value string) string`

Logic:
- Apply dark green foreground and restore default foreground.

Acceptance:
- Used for accents and added diff lines.

### `func greenBold(value string) string`

Logic:
- Apply dark green bold foreground and restore style.

Acceptance:
- Used for headings/accent labels.

### `func red(value string) string`

Logic:
- Apply red foreground and restore default foreground.

Acceptance:
- Used for tool errors and removed diff lines.

### `func dim(value string) string`

Logic:
- Apply dim gray foreground and restore default foreground.

Acceptance:
- Used for low-emphasis metadata.

### `func muted(value string) string`

Logic:
- Apply muted gray foreground and restore default foreground.

Acceptance:
- Used for context and footer text.

### `func ansiText(value string) string`

Logic:
- Apply near-white foreground and restore default foreground.

Acceptance:
- Used for normal assistant output.

### `func boldText(value string) string`

Logic:
- Apply near-white bold text.

Acceptance:
- Used for Markdown strong text and tool titles.

### `func italicText(value string) string`

Logic:
- Apply near-white italic text.

Acceptance:
- Used for Markdown emphasis.

### `func inlineCode(value string) string`

Logic:
- Apply muted yellow foreground for inline code.

Acceptance:
- Markdown inline code is visibly distinct.

### `func thinkingText(value string) string`

Logic:
- Apply dim gray italic text.

Acceptance:
- Thinking content is visually separate from final assistant text.

### `func thinkingStrong(value string) string`

Logic:
- Apply dim gray italic bold text.

Acceptance:
- Strong Markdown inside thinking remains in the thinking palette.

### `func userBackground(value string) string`

Logic:
- Apply user message background and restore default background.

Acceptance:
- User prompts render as Pi-like boxed blocks.

### `func toolPendingBackground(value string) string`

Logic:
- Apply pending tool background and restore default background.

Acceptance:
- Running tools can render separately from completed tools.

### `func toolSuccessBackground(value string) string`

Logic:
- Apply success tool background and restore default background.

Acceptance:
- Completed successful command/tool blocks are green.

### `func toolErrorBackground(value string) string`

Logic:
- Apply error tool background and restore default background.

Acceptance:
- Failed command/tool blocks are red.
