# `internal/tui/components/toolblock/display.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Extracts human-readable command/tool output.

## TODO

- [x] File exists in the split component architecture.
- [x] Bash, failed bash, and patch outputs are covered by tests.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash.

## Purpose

Turn raw tool arguments/results into the text shown inside a tool block.

## Functions

### `func (b *Block) formatContent() string`

Logic:
- Build title, optional argument JSON, and optional result output.

Acceptance:
- Bash command blocks suppress redundant raw argument JSON.

### `func (b *Block) title() (string, bool)`

Logic:
- Prefer `$ command` for bash, `tool path` for path-based tools, tool name,
  tool id, then a generic fallback.

Acceptance:
- Tool blocks have a useful first line even with incomplete metadata.

### `func (b *Block) output() string`

Logic:
- Decode JSON results and prefer patch, output, stdout/stderr, then pretty JSON.

Acceptance:
- Built-in tool results render useful content instead of raw escaped JSON.

### `func (b *Block) resultLooksFailed() bool`

Logic:
- Treat timed-out results and non-zero `exit_code` as failed command display.

Acceptance:
- Bash failures render red/error UI without changing agent control flow.

### `func stringField(raw string, key string) string`

Logic:
- Extract a string field from raw JSON.

Acceptance:
- Missing or invalid JSON returns empty string.

### `func numericField(values map[string]any, key string) (int, bool)`

Logic:
- Extract integer-like JSON fields from supported numeric forms.

Acceptance:
- `exit_code` detection works with standard JSON decoder values.

### `func prettyJSON(raw string) string`

Logic:
- Pretty print object JSON or return trimmed raw text.

Acceptance:
- Generic arguments remain readable.

### `func prettyObject(values map[string]any) string`

Logic:
- Marshal an object with indentation.

Acceptance:
- Empty objects render as empty output.

### `func decodeObject(raw string) (map[string]any, bool)`

Logic:
- Decode one JSON object using `json.Number`.

Acceptance:
- Non-object or invalid input returns false.
