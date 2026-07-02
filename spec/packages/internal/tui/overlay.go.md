# `internal/tui/overlay.go`

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

Manage overlay components and focus ownership.

## Code Style

Overlay stacking is explicit. Do not reuse disposed overlay handles.

## Functions

### `(*OverlayManager) Open(component Component, opts OverlayOptions) Handle`

Logic:

- Normalize configured paths and store injected dependencies.
- Apply defaults without touching the filesystem unless required by acceptance.
- Return a concrete value suitable for temp-directory tests.
- Positions by anchor or explicit row/col.
- Support visibility predicate.
- Focuses requested overlay.

Acceptance:

- positions by anchor or explicit row/col;
- supports visibility predicate;
- focuses requested overlay.

### `(*OverlayManager) Close(handle Handle)`

Logic:

- Look up the overlay by handle and return immediately if it is already closed or unknown.
- Remove it from z-order and invalidation indexes.
- Restore focus to the previous focus target when still available; otherwise choose the root component.
- Emit one invalidation for the old overlay bounds.

Acceptance:

- idempotently removes overlay and restores focus target.

Tests:

- `TestTUIOverlayStackFocus`
