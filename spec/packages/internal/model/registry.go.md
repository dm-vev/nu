# `internal/model/registry.go`

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

Build available model registry from built-ins, custom models, auth, settings,
and extensions.

## Code Style

Registry reads inputs and returns immutable snapshots. Do not mutate registry
while UI iterates it.

## Functions

### `BuildRegistry(inputs RegistryInputs) (*Registry, Diagnostics, error)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Merge built-ins, custom models, and extension models.
- Mark unavailable models when auth is missing.
- Custom provider/model entries override built-ins by id.

Acceptance:

- merges built-ins, custom models, and extension models;
- marks unavailable models when auth is missing;
- custom provider/model entries override built-ins by id.

### `(*Registry) Available() []Model`

Logic:

- Copy models marked available into a new slice.
- Sort by provider id, model display name, then model id for deterministic UI and tests.
- Return immutable copies so callers cannot mutate registry indexes.
- Exclude models hidden by auth, capability, or user settings filters.

Acceptance:

- returns stable sorted copy.

Tests:

- `TestNUF031UnavailableModelsHiddenWithoutAuth`
- `TestNUF031CustomModelsOverrideBuiltins`
