# `internal/tui`

## Status

Current: IN_PROGRESS
Implementation Commit: -

## Purpose

Nu's interactive terminal application root. It owns `App`, agent-event/prompt
orchestration, slash dispatch, and wiring across cohesive TUI subpackages. It
does not own concrete editor, engine, input, message, terminal, or component
implementations.

## Files

- root: app state, event/prompt/slash orchestration, and subpackage wiring;
- `core`: ANSI, layout primitives, component contracts, containers, and cursors;
- `editor`, `engine`, `input`, `message`, `terminal`: their named runtime layers;
- `components`: every reusable assistant/user/tool/markdown/menu/header/footer/
  box/text/status component in one cohesive package.

Files use normal names within each child (`decoder.go`, `render.go`,
`terminal.go`), not temporary prefixes such as `input_decoder.go` or
`terminal_terminal.go`.

## Acceptance Criteria

- `internal/tui` has exactly the seven approved child packages and no nested
  component packages.
- Agent events never write directly to stdout; they mutate state and render
  through `TUI`.
- Exit/abort paths restore raw terminal state.
- The component tree includes a filler before editor/footer so input stays at the bottom.
- Covered by `go test ./internal/tui/...`.
