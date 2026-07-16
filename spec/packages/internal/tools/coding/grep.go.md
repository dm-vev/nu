# `internal/tools/coding/grep.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Grep behavior is owned by `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Search cwd-contained files with literal or regexp matching.

## Code Style

Use stdlib walk/scanner/regexp, explicit scanner limits, and bounded line display.

## Owned Logic

- `RunGrep` validates arguments, resolves root, builds a matcher, walks non-ignored matching files, and returns sorted bounded matches.
- Matcher helpers implement literal/regexp and ignore-case modes.
- `grepFile` emits `path:line:text`, limits scanner tokens and match count, and truncates huge matching lines.

## Acceptance

- Literal, regexp, case, glob, ignore, result, and cwd limits are enforced.
- Invalid regex and scanner/file failures return contextual errors.

## Tests

- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`
- `TestGrepIgnoreCase`
- `TestGrepTruncatesLongMatchingLine`
- `TestGrepRejectsInvalidRegex`
