# `internal/config/trust.go`

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

Represent and persist project trust decisions.

## Code Style

No prompts here. This package decides from stored data and explicit overrides;
TUI/CLI asks the user elsewhere.

## Functions

### `ResolveTrust(ctx context.Context, store TrustStore, cwd string, req cli.Request, settings Settings) (TrustDecision, error)`

Logic:

- Normalize `cwd` and load stored trust records through `store`.
- Apply one-run CLI overrides first: `--approve` allows and `--no-approve` blocks project resources.
- For non-interactive mode, use `settings.defaultProjectTrust` without prompting.
- Match stored parent trust records against child cwd using cleaned path ancestry.

Acceptance:

- `--approve` trusts current project for this run;
- `--no-approve` blocks project resources for this run;
- non-interactive mode uses `defaultProjectTrust`;
- parent trust applies to child cwd.

### `SaveTrust(ctx context.Context, store TrustStore, decision TrustDecision) error`

Logic:

- Skip persistence for decisions marked one-run or CLI-only.
- Serialize project path, trust state, timestamp, and source in stable JSON.
- Write through `TrustStore` so tests avoid the real user configuration.
- Preserve unrelated trust records already present in the store.

Acceptance:

- writes stable JSON;
- does not save one-run CLI overrides unless explicitly requested.

Tests:

- `TestNUF011ApproveAllowsProjectResources`
- `TestNUF011NoApproveBlocksProjectResources`
