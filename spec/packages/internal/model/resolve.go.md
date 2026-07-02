# `internal/model/resolve.go`

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

Resolve provider/model patterns from CLI, settings, cycling, and slash commands.

## Code Style

Pure functions. Pattern matching must be deterministic and explain failures.

## Functions

### `Resolve(reg *Registry, selector Selector) (Model, error)`

Logic:

- Split optional `:thinking` suffix only after handling provider/model slash.
- If selector includes provider prefix, search only that provider.
- Match exact model id first, then exact model name, then glob patterns.
- Apply enabled-model list only when selector asks for cycle/default model.
- If one match remains, apply thinking override and return it.
- If zero matches, return a not-found error listing available providers.
- If multiple matches remain, return an ambiguity error with candidate ids.

Acceptance:

- supports `provider/model`, model id, model name, wildcard patterns, and
  optional `:thinking` suffix;
- errors on ambiguous matches.

### `Match(pattern string, model Model) bool`

Logic:

- Normalize provider/model/name comparison strings.
- Treat plain strings as exact id/name matches unless they contain glob
  metacharacters.
- For glob patterns, match against model id, model name, and `provider/model`.
- Return false for empty pattern.

Acceptance:

- supports simple glob semantics used by enabled model cycling.

Tests:

- `TestNUF031ModelPatternSelectsProviderAndModel`
