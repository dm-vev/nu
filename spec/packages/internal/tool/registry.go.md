# `internal/tool/registry.go`

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

Register, filter, and look up built-in and extension tools.

## Code Style

Return copies for lists. Duplicate names are explicit errors unless extension
override is allowed by future spec.

## Functions

### `NewRegistry(tools []Tool) (*Registry, error)`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Reject duplicate names.
- Index by exact tool name.

Acceptance:

- rejects duplicate names;
- indexes by exact tool name.

### `(*Registry) Filter(policy FilterPolicy) *Registry`

Logic:

- Start from the registry snapshot to avoid mutating the original registry.
- If `no-tools` is set, return an empty registry immediately.
- Remove built-ins when `no-builtin-tools` is set, then apply allowlist and exclude filters by exact name.
- Preserve deterministic ordering and tool ownership metadata in the filtered copy.

Acceptance:

- supports allowlist, exclude list, no-tools, and no-builtin-tools.

Tests:

- `TestToolRegistryRejectsDuplicates`
- `TestToolRegistryFilterPolicy`
