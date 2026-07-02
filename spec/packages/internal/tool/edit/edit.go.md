# `internal/tool/edit/edit.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 6ec7970
Implementation Comments: Edit tool lives in its own subpackage with exact replacement and CRLF tests.

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
- Apply replacements to the editable content.
- Write final content.
- Return JSON with `path` and patch text.

Acceptance:

- applies a single exact replacement;
- rejects missing or ambiguous old text;
- preserves existing CRLF line endings unless replacement changes them.

Tests:

- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`
- `TestEditRejectsMissingOldText`
