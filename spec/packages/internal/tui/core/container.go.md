# `internal/tui/core/container.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.
Implementation Comments: Implement ordered child composition.

## TODO

- [x] File exists in the split `internal/tui` architecture.
- [x] Implementation is covered by at least one package-level TUI test path.
- [x] Current status is recorded in this spec file.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement ordered child composition.

## Code Style

Keep interfaces tiny and structural. Containers own ordering only, not layout policy. Add comments only for non-trivial control flow, terminal side effects, or state invariants.

## Acceptance Criteria

- File status is kept current before implementation commit.
- Behavior is covered by `go test ./internal/tui/...`.

## Types And Constants

### `type Container struct {`

Logic:
- Container renders children in order.

Acceptance:
- Used only inside the package boundary unless exported by current API needs.

## Functions

### `func (c *Container) AddChild(component Component)`

Logic:
- AddChild appends a component.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (c *Container) RemoveChild(component Component)`

Logic:
- RemoveChild removes a component by identity.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (c *Container) Clear()`

Logic:
- Clear removes all children.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (c *Container) Invalidate()`

Logic:
- Invalidate clears child caches.

Acceptance:
- Covered by the package tests and `go test ./internal/tui/...`.

### `func (c *Container) Render(width int) []string`

Logic:
- Render renders every child at the same width.

Acceptance:
- ANSI-stripped output never exceeds the requested width and repaint does not append duplicate full frames.
