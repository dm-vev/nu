# `internal/model/thinking.go`

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

Represent thinking levels and provider mappings.

## Code Style

No stringly-typed thinking outside this package.

## Functions

### `ParseThinkingLevel(value string) (ThinkingLevel, bool)`

Logic:

- Normalize the input by trimming whitespace and lowercasing ASCII letters.
- Match only the configured thinking level tokens.
- Return `false` without fallback when the token is empty or unknown.
- Do not map the level to provider-specific fields in this parser.

Acceptance:

- accepts only supported levels.

### `MapThinking(model Model, level ThinkingLevel) (ProviderThinking, error)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Apply model-specific thinking maps.
- Report unsupported levels with provider/model context.

Acceptance:

- applies model-specific thinking maps;
- reports unsupported levels with provider/model context.

Tests:

- `TestNUF032ThinkingLevelMapping`
- `TestNUF032UnsupportedThinkingLevelFallsBackOrErrors`
