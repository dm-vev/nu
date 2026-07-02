# `internal/message/usage.go`

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

Represent token usage and cost.

## Code Style

Use integer token counts and decimal-safe cost representation when precision is
needed. Avoid floats in persisted totals if they cause unstable JSON.

## Functions

### `AddUsage(a, b Usage) Usage`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Sum input, output, cache, total, and cost fields.
- Handle zero values.

Acceptance:

- sums input, output, cache, total, and cost fields;
- handles zero values.

### `EstimateTokens(messages []Message) int`

Logic:

- Perform a deterministic pure computation from the provided inputs.
- Return structured output that callers can test without external state.
- Deterministic estimate used for compaction thresholds.
- Never returns negative values.

Acceptance:

- deterministic estimate used for compaction thresholds;
- never returns negative values.

Tests:

- `TestUsageAdd`
- `TestUsageEstimateDeterministic`
