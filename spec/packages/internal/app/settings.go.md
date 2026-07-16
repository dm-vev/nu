# `internal/app/settings.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Provider settings schema, loading, and atomic persistence were split from runtime/model selection.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Read and persist Nu provider settings and the selected default model.

## Code Style

Use injected home paths, stdlib JSON/filesystem calls, and temp-file replacement.

## Owned Logic

- `providerSetting` and `providerSettingsFile` define global and per-provider configuration.
- `loadProviderSettings` returns empty defaults for no home or a missing file and normalizes a nil providers map.
- `saveSelectedModel` updates global and provider defaults without discarding other settings.
- `writeProviderSettings` writes indented JSON through a same-directory temp file and rename.

## Acceptance

- Missing settings are not errors.
- Selecting a model preserves unrelated provider configuration.
- Failed writes do not intentionally replace the existing settings file.

## Tests

- `TestPrintModeBuildsProviderFromSettings`
- `TestSavedModelSelectionRestoresDefault`
