# `internal/compact/branch.go`

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

Branch summarization when tree navigation leaves work behind.

## Code Style

Common-ancestor and entry collection are pure functions.

## Functions

### `PlanBranchSummary(tree *session.Tree, fromID, toID string, budget Budget) (BranchPlan, error)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Find deepest common ancestor.
- Collect abandoned branch entries newest-first until budget.
- Track file reads/writes cumulatively.

Acceptance:

- finds deepest common ancestor;
- collects abandoned branch entries newest-first until budget;
- tracks file reads/writes cumulatively.

### `SummarizeBranch(ctx context.Context, input BranchInput) (session.Entry, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Write branch summary entry at target position.
- Include source branch id.

Acceptance:

- writes branch summary entry at target position;
- includes source branch id.

Tests:

- `TestNUF091BranchSummaryFindsCommonAncestor`
- `TestNUF091BranchSummaryTracksFiles`
