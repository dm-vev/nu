# `internal/tui/status_status.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Transient status-line component.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Transient status-line component.

## Code Style

Components render `[]string` at supplied width and must not write to the terminal directly. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Rendered visible widths are bounded by the supplied width.

## Types And Constants

### `type Status struct {`

Logic:
- Status renders transient agent state.
- It keeps the current text, style callback, animation frame index, alert mode, and frame set.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func NewStatus(style func(string) string, frameSet ...string) *Status`

Logic:
- New creates an empty status line.
- Use Claude-like Unicode frames by default, or the supplied frame set when limited terminals require ASCII.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (s *Status) SetText(value string)`

Logic:
- SetText replaces status text.
- Reset the animation frame when the label changes so new states start from a stable visual point.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (s *Status) Text() string`

Logic:
- Text returns the raw status text.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func (s *Status) Step()`

Logic:
- Advance the Claude-like icon animation used while the agent is active.

Acceptance:
- `TestStatusStatusStepAnimatesLabel` fails if status frames stop changing.

### `var statusDefaultFrames`

Logic:
- Store deterministic Claude-like Unicode spinner frames: interpunct and asterisk variants.

Acceptance:
- Frames are stable and do not change visible row count.

### `func (s *Status) SetAlertText(value string)`

Logic:
- Replace status text and enable alert rendering for retry/error states.

Acceptance:
- Rate-limit retry status can use the same row while visually shifting toward red.

### `func (s *Status) Invalidate() {}`

Logic:
- Invalidate exists for the component interface.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.

### `func statusIdentity(value string) string`

Logic:
- Return the provided value unchanged as the default status style callback.

Acceptance:
- Covered by the package tests and `go test ./internal/tui`.
