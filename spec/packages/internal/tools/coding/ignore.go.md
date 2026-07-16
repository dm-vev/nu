# `internal/tools/coding/ignore.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Coding glob and root `.gitignore` helpers live in `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Provide the minimal ignore/glob behavior shared by find and grep.

## Code Style

Keep stdlib-only root-pattern matching; do not claim full gitignore semantics.

## Owned Logic

- `loadGitignore` reads non-empty, non-comment root patterns.
- `shouldSkip` always skips `.git` and matches directory, relative, or base-name patterns.
- `globMatches`/`wildcardMatch` apply optional `filepath.Match` patterns safely.

## Acceptance

- Find and grep skip `.git` and tested root ignore patterns.
- Empty globs match all files and invalid globs do not match.

## Tests

- `TestNUF074GrepRespectsGitignore`
- `TestNUF075FindRespectsGitignore`
