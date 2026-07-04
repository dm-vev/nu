# `internal/tui`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Top-level Nu interactive wiring. This package owns `App`, agent event handling,
prompt submission, raw/line input loop selection, and composition of the
Pi-style component tree. It must stay thin; rendering, terminal IO, input
decoding, editor mutation, and reusable components live in subpackages.

## Files

- `app.go`: app state and construction.
- `options.go`: external option normalization.
- `layout.go`: fixed component tree and component option construction.
- `events.go`: agent event to UI state mapping.
- `messages.go`: structured chat message component rebuilds.
- `run.go`: terminal lifecycle and raw/line loops.
- `submit.go`: prompt submission, abort, and write-error handling.
- `style.go`: built-in green/gray/white palette.
- `git.go`: branch discovery.
- `util.go`: small app helpers.
- `app_test.go`: top-level app rendering/input tests.

## Acceptance Criteria

- No reusable renderer/editor/terminal logic is implemented in the top-level package.
- Agent events never write directly to stdout; they mutate state and call `engine.TUI`.
- Exit/abort paths restore raw terminal state.
- The component tree includes a filler before editor/footer so input stays at the bottom.
- Covered by `go test ./internal/tui`.
