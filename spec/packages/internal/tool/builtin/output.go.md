# `internal/tool/builtin/output.go`

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

Streaming output accumulator for bash and future process tools.

## Code Style

Bound memory. Preserve UTF-8 validity. Temp file creation is injected for tests.

## Functions

### `NewOutputAccumulator(opts OutputOptions) *OutputAccumulator`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Start in memory-only mode.

Acceptance:

- starts in memory-only mode.

### `(*OutputAccumulator) Append(data []byte) error`

Logic:

- Feed bytes through a streaming UTF-8 decoder and hold incomplete runes until more data arrives.
- Append decoded text to the in-memory display buffer until display limits are exceeded.
- Track total bytes, total lines, and whether output has been truncated.
- Create and append to a temp spill file when full output must be preserved.

Acceptance:

- decodes streaming UTF-8;
- tracks total bytes and lines;
- spills full output to temp file when truncated/persisted.

### `(*OutputAccumulator) Snapshot(persistIfTruncated bool) OutputSnapshot`

Logic:

- Finalize any pending UTF-8 decoder state into replacement-safe display text.
- Build display output using configured head/tail policy and truncation metadata.
- If requested and truncated, ensure full output has a persisted path.
- Return display text, totals, truncation reason, and optional full-output path.

Acceptance:

- returns truncated display content plus full output path when persisted.

Tests:

- `TestOutputAccumulatorKeepsTail`
- `TestOutputAccumulatorPersistsFullOutput`
