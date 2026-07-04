# `internal/tui/components/footer`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Two-line footer showing cwd/branch and context/provider/model state.

## Files

- `options.go`: footer data.
- `footer.go`: component state.
- `format.go`: path/token/model formatting.
- `render.go`: width-bounded line alignment.
- `footer_test.go`: path, stats, model, and width checks.

## Acceptance Criteria

- Home paths are shortened to `~`.
- Provider/model display name uses configured model label when supplied.
- Each line renders at exactly the requested visible width.
