# `internal/compact/serialize.go`

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

Serialize conversations into summarizer prompts.

## Code Style

Pure text conversion. No token counting here.

## Functions

### `SerializeConversation(entries []session.Entry) string`

Logic:

- Iterate entries in active conversation order and emit stable section labels for
  user, assistant, thinking, tool calls, tool results, bash execution, custom
  content, compaction summaries, and branch summaries.
- Include tool names, ids, exit status, truncation metadata, and file references
  needed by the summarizer.
- Replace binary image payloads with image markers that preserve MIME type,
  dimensions when known, and source path/reference.
- Keep output deterministic: no timestamps generated here, no token counting,
  no provider calls.

Acceptance:

- labels user, assistant, thinking, tool calls, tool results, bash execution,
  custom, compaction, and branch summary content;
- omits binary image data while preserving image markers.

### `ExtractFileTracking(entries []session.Entry) FileTracking`

Logic:

- Walk tool result details, bash execution metadata, branch summaries, and
  previous compaction summaries in entry order.
- Add read paths, written paths, and deleted paths into separate stable sets.
- Normalize paths exactly as stored by the tool layer; do not resolve them
  against the current process cwd.
- Return sorted lists so compaction output is deterministic.

Acceptance:

- accumulates read and modified files from tool calls and previous summaries.

Tests:

- `TestCompactSerializeConversation`
- `TestCompactExtractFileTracking`
