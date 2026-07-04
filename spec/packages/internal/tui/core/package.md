# `internal/tui/core`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Minimal component contracts shared by the TUI packages.

## Files

- `component.go`: render/invalidate/focus interfaces.
- `container.go`: ordered child composition.
- `cursor.go`: cursor marker and cursor coordinates.
- `line.go`: per-line reset application.
- `container_test.go`: child ordering test.

## Acceptance Criteria

- Interfaces remain structural and tiny.
- Containers only compose children; they do not own layout policy.
- No terminal IO is allowed in this package.
