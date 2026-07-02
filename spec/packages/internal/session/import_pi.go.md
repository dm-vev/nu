# `internal/session/import_pi.go`

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

Import important Pi session JSONL records into Nu's native session format.

## Code Style

Importer is best-effort but explicit. It never mutates Pi files.

## Functions

### `ImportPiSession(ctx context.Context, r io.Reader, opts ImportOptions) (*Session, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Stream or decode records in order without mutating the source.
- Preserve unsupported data as diagnostics/details when possible.
- Maps Pi message roles and content blocks to Nu messages.
- Preserve unknown Pi data in details where useful.
- Report unsupported entry types as diagnostics, not silent loss.

Acceptance:

- maps Pi message roles and content blocks to Nu messages;
- preserves unknown Pi data in details where useful;
- reports unsupported entry types as diagnostics, not silent loss.

Tests:

- `TestPiImportBasicMessages`
- `TestPiImportPreservesToolResults`
