# `internal/tool/builtin/edit.go`

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

Built-in `edit` tool.

## Code Style

Keep match/patch generation pure and separate from filesystem writes.

## Functions

### `NewEdit(cwd string, opts EditOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `edit`, schema, and sequential mutation behavior.

Acceptance:

- registers name `edit`, schema, and sequential mutation behavior.

### `ApplyEdits(original string, edits []Edit) (EditResult, error)`

Logic:

- Detect original file line ending style before normalization.
- Normalize original content and each edit text to LF for matching.
- For every edit, find all occurrences of `oldText` in the original normalized
  content, not incrementally mutated content.
- Reject edits whose `oldText` has zero matches or more than one match.
- Convert each match to a byte/rune span and sort spans by start offset.
- Reject overlapping or nested spans.
- Apply replacements from the end of the file toward the beginning so offsets
  remain valid.
- Restore original line endings in the final content.
- Generate display diff, unified patch, and first changed line from original vs
  final content.

Acceptance:

- matches edits against original content;
- rejects missing, ambiguous, and overlapping old text;
- preserves CRLF/LF line endings.

### `executeEdit(ctx context.Context, args EditArgs, ops EditOps) tool.Result`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Resolve target path against cwd and check read/write access.
- Run the full read/apply/write sequence inside per-file mutation queue.
- Read file bytes and preserve UTF-8 BOM if present.
- Call `ApplyEdits` before writing anything.
- Write final UTF-8 content only after patch generation succeeds.
- Return tool details containing display diff, unified patch, and first changed
  line.

Acceptance:

- reads file, applies edits, writes file, returns unified patch.

Tests:

- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`
