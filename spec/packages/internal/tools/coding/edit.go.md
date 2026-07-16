# `internal/tools/coding/edit.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Exact-edit behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Apply unambiguous exact replacements to one cwd-contained file.

## Code Style

Use original byte spans and stdlib string operations; do not add a diff dependency.

## Owned Logic

- `RunEdit` validates replacements/path/context, serializes mutation, and reads the original file.
- It requires each old text exactly once, rejects overlaps, applies sorted spans in reverse, writes the result, and returns a small patch string.

## Acceptance

- Multiple replacements are matched against original content.
- Missing, ambiguous, empty, or overlapping replacements fail before write.
- Existing CRLF is preserved unless replacement content changes it.

## Tests

- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`
- `TestEditAppliesMultipleReplacementsAgainstOriginal`
