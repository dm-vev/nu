# `internal/session/types.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Shared store/session/header types and resolution/import errors were split from store implementation.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Define shared session-store value types, errors, limits, and synchronization state.

## Code Style

Keep this file declarative; behavior belongs in its domain file.

## Owned Logic

- Errors distinguish not-found, ambiguous selector, and oversized import cases.
- `Ref` identifies an ID with optional explicit path and cwd.
- `Store`, `Session`, and `Header` define store state and persisted header shape.
- `maxImportBytes` bounds untrusted imports at 32 MiB.

## Acceptance

- Callers can classify selector/import failures with `errors.Is`.
- Header fields match `spec/protocols/session-jsonl.md`.
- Stores retain no global path or lock state.

## Tests

- `TestNUF081ResumeByPathOrPartialID`
- `TestSessionImportRejectsOversizedInput`
- `TestNUF080SessionAppendBuildsTree`
