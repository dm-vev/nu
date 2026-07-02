# `internal/model/model.go`

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

Define provider/model metadata.

## Code Style

Types only. Keep provider request behavior out of model metadata.

## Types

### `Model`, `Provider`, `Cost`, `Capabilities`

Logic:

- Keep these as plain metadata structs with no registry, auth, HTTP, or pricing
  lookup behavior.
- Represent the provider id, model id, display name, API kind, input content
  types, context window, output token limit, pricing, compatibility flags, and
  thinking controls.
- Use zero values only for truly unknown metadata; defaults are applied by
  registry construction, not by struct methods.
- Keep vendor-specific capabilities extensible without forcing provider request
  conversion into this package.

Acceptance:

- represents provider id, model id/name, API kind, input types, context window,
  max tokens, cost, compat, and thinking support.

## Functions

No matching or registry functions belong in this file. Matching stays in
`resolve.go`; registry construction stays in `registry.go`.

Tests:

- covered by registry and resolver tests.
