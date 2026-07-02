# `internal/model/model.go`

## Status

Current: IN_PROGRESS
Implementation Commit: 4ddd508
Implementation Comments: Phase 3 registry covers built-in/custom models, auth filtering, explicit disabled custom models, thinking mappings, and is being extended with configurable display names.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Represent model metadata, available-model filtering, pattern matching, custom
model overrides, and provider thinking-level mapping.

## Code Style

One small package, no network calls. Use explicit structs instead of dynamic
maps so CLI/list-models callers get stable fields.

## Types

### `type Model`

Logic:

- Store id, provider, API kind, optional display name, aliases, patterns, enabled flag,
  requires-auth flag, capabilities, context window, max output, input/output
  cost per million tokens, and thinking support.

Acceptance:

- enough metadata exists to select and list models without provider calls.

### `type Registry`

Logic:

- Hold built-in and custom models by provider/id.
- Preserve deterministic ordering for list output.

Acceptance:

- custom models override matching built-ins.

## Functions

### `Builtins() []Model`

Logic:

- Return representative built-in models for OpenAI Responses, OpenAI Chat,
  Anthropic Messages, Google GenerateContent, and Bedrock ConverseStream.
- Mark provider-auth-required models.

Acceptance:

- built-ins cover every Phase 3 provider adapter.

### `NewRegistry(models []Model) Registry`

Logic:

- Normalize enabled default to true.
- Index models by `provider/id`.
- Keep stable sorted keys.

Acceptance:

- duplicate keys are overridden by the last model.

### `LoadCustom(path string) ([]Model, error)`

Logic:

- Return nil for missing files.
- Decode `{ "models": [...] }`.
- Validate provider, API, and id.
- Preserve explicit `enabled: false` and default omitted enabled to true.

Acceptance:

- custom models can override built-ins.

### `Resolve(pattern string, auth map[string]bool) (Model, error)`

Logic:

- Match exact id, `provider/id`, aliases, and glob patterns.
- Hide disabled models and auth-required models when provider auth is absent.

Acceptance:

- provider/model patterns select the expected model;
- unavailable models are hidden without auth.

### `Available(auth map[string]bool) []Model`

Logic:

- Return enabled models visible under the auth state.
- Sort by provider then id.

Acceptance:

- output is deterministic.

### `ThinkingFor(providerID string, api string, level ThinkingLevel) (map[string]any, error)`

Logic:

- Translate `off`, `minimal`, `low`, `medium`, `high`, and `xhigh` into
  provider request fragments.
- Return nil for `off`.
- Error when a provider/API cannot support a non-off level.

Acceptance:

- supported levels map to provider-specific fields;
- unsupported levels fail clearly.

Tests:

- `TestNUF031ModelPatternSelectsProviderAndModel`
- `TestNUF031UnavailableModelsHiddenWithoutAuth`
- `TestNUF031CustomModelsOverrideBuiltins`
- `TestNUF031CustomModelDisplayNameLoads`
- `TestCustomModelsCanDisableEntry`
- `TestNUF032ThinkingLevelMapping`
- `TestNUF032UnsupportedThinkingLevelFallsBackOrErrors`
