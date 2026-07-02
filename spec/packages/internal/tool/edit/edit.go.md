# `internal/tool/edit/edit.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Edit tool lives in its own subpackage with exact replacement, CRLF, and original-span multi-replacement tests.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `edit` built-in tool.

## Code Style

Exact replacement only. No diff library. Return a small patch string that proves
the change without pretending to be a full patch engine.

## Functions

### `Run(ctx context.Context, cwd string, raw string) (agent.ToolResult, error)`

Logic:

- Decode `path` and non-empty `replacements`.
- Resolve path under cwd.
- Check cancellation before mutation.
- Acquire `toolkit.MutationMu`.
- Read original content.
- For each replacement, require non-empty `old` and exactly one match in the
  original content.
- Record each replacement as a byte span in the original content.
- Sort spans and reject overlaps.
- Apply replacements from the end of the file back to the front, so one
  replacement cannot affect another replacement's target.
- Write final content.
- Return JSON with `path` and patch text.

Acceptance:

- applies a single exact replacement;
- applies multiple replacements against original content, not against prior
  replacement output;
- rejects missing or ambiguous old text;
- rejects overlapping replacements;
- preserves existing CRLF line endings unless replacement changes them.

Tests:

- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`
- `TestEditAppliesMultipleReplacementsAgainstOriginal`
- `TestEditRejectsMissingOldText`
