# `internal/tools/coding/find.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Find behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Find cwd-contained files by optional root/glob with ignore and count limits.

## Code Style

Use stdlib `WalkDir` and return sorted slash-form relative paths.

## Owned Logic

- `RunFind` defaults root to `.` and limit to 100, resolves containment, loads root ignore rules, and walks files.
- It checks cancellation, skips ignored paths, matches relative/base names, stops at the limit, sorts, and returns bounded JSON.

## Acceptance

- Glob, `.gitignore`, cancellation, count, output-byte, and cwd boundaries are enforced.

## Tests

- `TestNUF075FindGlob`
- `TestNUF075FindRespectsGitignore`
- `TestFindEnforcesLimit`
- `TestFindRejectsPathEscape`
