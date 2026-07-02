# `internal/compact/compact.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 2931429
Implementation Comments: Phase 4 compaction planning and branch-summary metadata are deterministic pure functions over session entries; oversize single entries are compacted instead of exceeding budget.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Small deterministic helpers for Phase 4 compaction planning and branch-summary
metadata. No provider calls and no filesystem access.

## Code Style

Use plain slices and maps. Do not introduce a summarizer interface until a real
LLM compaction call exists.

## Types

### `type Plan struct`

Logic:

- Store entries kept in prompt context.
- Store entries selected for compaction.
- Store the cut index used to split the original path.

Acceptance:

- callers can tell exactly what remains and what needs summarizing.

### `type BranchSummary struct`

Logic:

- Store common ancestor id.
- Store abandoned branch entry ids.
- Store touched file paths found in abandoned payload metadata.

Acceptance:

- callers can persist branch-summary metadata without rereading the paths.

## Functions

### `BuildPlan(entries []session.Entry, contextWindow int, reserveTokens int) Plan`

Logic:

- Estimate each entry cost from `payload.tokens` when present, otherwise from
  payload JSON size.
- Keep all entries when the total estimate fits `contextWindow - reserveTokens`.
- Otherwise keep the newest suffix that fits the budget.
- If no entry can fit the budget, return an empty keep slice and compact all entries.
- If the suffix starts with a tool-result message, include the preceding
  assistant tool-call message with the same `tool_call_id`.

Acceptance:

- recent messages remain under budget;
- oversized entries do not exceed the keep budget;
- compaction never keeps a tool result without its matching tool call.

### `BuildBranchSummary(from []session.Entry, to []session.Entry) BranchSummary`

Logic:

- Find the last common entry id between two root-to-leaf paths.
- Mark entries in `from` after the ancestor as abandoned.
- Collect unique file paths from `payload.files` arrays in abandoned entries.
- Preserve file order by first appearance.

Acceptance:

- common ancestor is deterministic;
- file tracking is stable and duplicate-free.

Tests:

- `TestNUF090CompactionKeepsRecentBudget`
- `TestCompactionCompactsOversizedSingleEntry`
- `TestNUF090CompactionDoesNotCutBeforeToolResult`
- `TestNUF091BranchSummaryFindsCommonAncestor`
- `TestNUF091BranchSummaryTracksFiles`
