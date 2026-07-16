# `internal/session/compaction.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Deterministic compaction planning moved from the removed standalone package into `internal/session`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Split a session path into recent entries to keep and older entries to compact.

## Code Style

Keep planning pure and deterministic; do not add a summarizer abstraction before a real provider-backed summary exists.

## Owned Logic

- `Plan` stores kept/compacted slices and their cut index.
- `BuildPlan` computes available budget, keeps the newest fitting suffix, and expands the suffix to include a matching assistant tool call before a tool result.
- Metadata helpers use explicit `payload.tokens` or a minimum-one byte estimate fallback.

## Acceptance

- All entries remain when they fit; oversized entries can leave an empty keep slice.
- Recent context stays within the estimated budget.
- A kept tool result is not separated from its matching tool call.

## Tests

- `TestNUF090CompactionKeepsRecentBudget`
- `TestCompactionCompactsOversizedSingleEntry`
- `TestNUF090CompactionDoesNotCutBeforeToolResult`
