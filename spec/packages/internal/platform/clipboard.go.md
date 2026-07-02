# `internal/platform/clipboard.go`

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

Clipboard text/image integration.

## Code Style

Optional platform feature. Fail gracefully when unavailable.

## Functions

### `ReadImage(ctx context.Context) (Image, bool, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Return `ok=false` when clipboard has no supported image.
- Never blocks indefinitely.

Acceptance:

- returns `ok=false` when clipboard has no supported image;
- never blocks indefinitely.

### `WriteText(ctx context.Context, text string) error`

Logic:

- Check context before invoking platform clipboard commands or APIs.
- Select the platform writer configured for the current build target.
- Write text as UTF-8 without adding a newline.
- Return a typed unsupported error when no clipboard writer is available.

Acceptance:

- writes text or returns unsupported error.

Tests:

- `TestClipboardUnsupportedGraceful`
