# `internal/compact/compact.go`

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

Auto and manual context compaction.

## Code Style

Cut-point logic is pure and heavily tested. Provider call for summarization is
separate from selection.

## Functions

### `Plan(messages []session.Entry, budget Budget) (PlanResult, error)`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Keep recent messages within budget.
- Do not cut between tool call and tool result.
- Handle split-turn case explicitly.

Acceptance:

- keeps recent messages within budget;
- does not cut between tool call and tool result;
- handles split-turn case explicitly.

### `Compact(ctx context.Context, input Input) (session.Entry, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Serializes selected messages.
- Call summarizer provider.
- Return compaction entry with first kept id and file tracking.

Acceptance:

- serializes selected messages;
- calls summarizer provider;
- returns compaction entry with first kept id and file tracking.

Tests:

- `TestNUF090CompactionKeepsRecentBudget`
- `TestNUF090CompactionDoesNotCutBeforeToolResult`
