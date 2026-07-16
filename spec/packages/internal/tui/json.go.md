# `internal/tui/json.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Shared trust/auth JSON mutation moved from the former aggregate slash-handler file.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Merge and persist the small JSON files used by slash auth/trust handlers.

## Code Style

Use stdlib JSON/filesystem calls, preserve unrelated map entries, and write mode `0600`.

## Owned Logic

- `writeBoolMap` merges one trust key.
- `writeAuthProvider` merges one provider credential and `removeAuthProvider` deletes one provider.
- `writeJSONFile` creates parents and writes indented JSON plus LF with contextual errors.

## Acceptance

- Trust/login/logout updates do not delete unrelated entries.
- Missing files initialize valid shapes.
- Write and encoding failures include the target path.

## Tests

- `TestTUIAllBuiltinSlashCommandsHaveHandlers`
- `TestTUISlashSessionDoesNotCallAgent`
