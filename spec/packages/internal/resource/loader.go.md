# `internal/resource/loader.go`

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

Discover all resources allowed for a run.

## Code Style

Resource loading is deterministic and returns diagnostics instead of printing.

## Functions

### `Load(ctx context.Context, opts Options) (Set, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Load context files, skills, prompts, themes, packages, and extensions.
- Apply trust and CLI disable flags.
- Preserve source information for UI diagnostics.

Acceptance:

- loads context files, skills, prompts, themes, packages, and extensions;
- applies trust and CLI disable flags;
- preserves source information for UI diagnostics.

Tests:

- `TestResourceLoaderRespectsTrust`
- `TestResourceLoaderAppliesDisableFlags`
