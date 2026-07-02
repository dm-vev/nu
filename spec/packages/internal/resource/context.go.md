# `internal/resource/context.go`

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

Load context files and system prompt files.

## Code Style

Filesystem walking is explicit and bounded. No hidden reads from real home in
tests.

## Functions

### `LoadContextFiles(ctx context.Context, opts ContextOptions) ([]ContextFile, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Walk parent directories for `AGENTS.md` and `CLAUDE.md`.
- Load global and project context in precedence order.
- Support `--no-context-files`.

Acceptance:

- walks parent directories for `AGENTS.md` and `CLAUDE.md`;
- loads global and project context in precedence order;
- supports `--no-context-files`.

### `LoadSystemPrompt(ctx context.Context, opts ContextOptions) (SystemPrompt, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Support replacement and append prompt files.

Acceptance:

- supports replacement and append prompt files.

Tests:

- `TestNUF120ContextFilesWalkParents`
- `TestNUF120SystemPromptReplacementAndAppend`
