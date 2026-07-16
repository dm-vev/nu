# `internal/tools/coding/result.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Coding-tool JSON/result and image helpers live in `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Define Nu tool results and shared JSON/output formatting helpers.

## Code Style

Keep helpers deterministic, stdlib-only, and independent of filesystem mutation.

## Owned Logic

- `Result` carries SDK-adaptable JSON content.
- `decodeArgs` treats blank input as `{}` and wraps JSON errors.
- Truncation and JSON helpers bound strings/lists while preserving valid JSON and empty arrays.
- `imageMIME` recognizes PNG, JPEG, GIF, and WebP extensions.

## Acceptance

- Every built-in returns valid JSON content.
- Bounded lists drop trailing values and set `truncated`.
- Empty lists encode as `[]`, not `null`.

## Tests

- Covered by `TestBuiltinsReadToolRuns` and read/bash/grep/find/ls package tests.
