# `internal/tools/coding/path.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Cwd path sandboxing lives in `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Resolve tool paths beneath cwd while blocking lexical and symlink escapes.

## Code Style

Use stdlib path operations and prove containment through the nearest existing parent for create targets.

## Owned Logic

- `resolveUnder` requires relative paths, resolves cwd/path symlinks, proves containment, and returns slash-form relative paths.
- `resolveForContainment` permits missing leaves only after resolving an existing parent.
- Helpers test containment, derive relative paths, and apply path defaults.

## Acceptance

- Absolute, `..`, existing symlink, and missing-target symlink-parent escapes fail.
- `.` resolves to cwd and new contained files are allowed.

## Tests

- `TestReadRejectsPathEscape`
- `TestReadRejectsSymlinkEscape`
- `TestWriteRejectsSymlinkParentEscape`
- `TestFindRejectsPathEscape`
