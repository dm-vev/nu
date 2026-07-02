# `internal/tui/component.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Base component interfaces and simple layout primitives.

## Code Style

Small interfaces. Components render to lines and handle input only when focused.

## Types

### `Component`, `Focusable`, `Container`, `Text`, `Box`, `Spacer`

Logic:

- Define small rendering interfaces around size, render, focus, and invalidation behavior.
- Require every component render to respect the width/height passed by parent layout.
- Keep cached render state invalidated by explicit calls, not by hidden global terminal state.
- Propagate focus through containers to the active focusable child.

Acceptance:

- components render within requested width;
- invalidation clears cached render state;
- focus state propagates to embedded focusable children.

## Functions

Component constructors for primitives may live here only when they are trivial:
`NewContainer`, `NewText`, `NewBox`, and `NewSpacer`. Complex widgets get their
own file spec.

Tests:

- `TestTUIContainerFocusPropagation`
