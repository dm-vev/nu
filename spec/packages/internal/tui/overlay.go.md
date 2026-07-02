# `internal/tui/overlay.go`

## Status

Current: IMPLEMENTED
Implementation Commit: pending
Implementation Comments: Overlay stack uses small comparable handles, restores previous focus on close, and rejects disposed handles.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Track focused overlay stack behavior independent from rendering.

## Code Style

Use small concrete structs. Comment focus restoration and disposed-handle
guards.

## Types

### `type OverlayStack struct`

Logic:

- Store active overlay records in focus order.
- Track disposed handles so they cannot be closed or reused twice.

Acceptance:

- focused visible overlay receives input first.

### `type OverlayHandle struct`

Logic:

- Carry immutable id and title for caller-owned close requests.

Acceptance:

- handles remain comparable by id.

## Functions

### `NewOverlayStack() *OverlayStack`

Logic:

- Return an empty stack.

Acceptance:

- focused overlay is empty.

### `(*OverlayStack) Push(title string) OverlayHandle`

Logic:

- Create a new handle id.
- Push it as the focused overlay.

Acceptance:

- newest overlay is focused.

### `(*OverlayStack) Close(handle OverlayHandle) bool`

Logic:

- Reject disposed or unknown handles.
- Remove the matching overlay.
- Restore focus to the previous still-active overlay.
- Mark the handle disposed.

Acceptance:

- closing focused overlay restores previous focus;
- disposed handles cannot be reused.

### `(*OverlayStack) Focused() (OverlayHandle, bool)`

Logic:

- Return the current top overlay when present.

Acceptance:

- returns false for an empty stack.

Tests:

- `TestNUF100OverlayFocusRestoresPrevious`
