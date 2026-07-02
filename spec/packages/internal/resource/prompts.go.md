# `internal/resource/prompts.go`

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

Load and expand prompt templates.

## Code Style

Template expansion is pure. Do not execute shell or arbitrary code.

## Functions

### `DiscoverPrompts(ctx context.Context, opts PromptOptions) ([]PromptTemplate, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Load Markdown templates from allowed locations.
- Parses description and argument hint frontmatter.

Acceptance:

- loads Markdown templates from allowed locations;
- parses description and argument hint frontmatter.

### `ExpandPrompt(t PromptTemplate, args []string) (string, error)`

Logic:

- Scan the template left-to-right and copy literal text unchanged.
- Expand `$1`, `$@`, and `$ARGUMENTS` from the provided args without shell evaluation.
- Handle `${1:-default}`, `${@:N}`, and `${@:N:L}` using one-based positional argument rules.
- Return a syntax error with template name/offset for malformed placeholders.

Acceptance:

- supports `$1`, `$@`, `$ARGUMENTS`, `${1:-default}`, `${@:N}`, and `${@:N:L}`.

Tests:

- `TestNUF140PromptTemplateExpansion`
- `TestNUF140PromptTemplateDefaultArgument`
