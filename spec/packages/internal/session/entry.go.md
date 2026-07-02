# `internal/session/entry.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 2931429
Implementation Comments: Basic JSONL entry envelope exists; Phase 4 preserves unknown top-level fields and uses raw payload helpers for session-state/compaction metadata.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define persisted session JSONL entry schema.
Implements `spec/protocols/session-jsonl.md`.

## Code Style

Versioned tagged records. Unknown future fields are preserved where practical.

## Types

### `Entry`

Logic:

- Represent the common entry envelope from `spec/protocols/session-jsonl.md`.
- Keep `payload` as raw JSON until typed message storage is implemented.
- Preserve unknown optional top-level fields in `extra` during import/migration.
- Use `parent_id` exactly as persisted; tree validation belongs to `tree.go`.

Acceptance:

- includes schema version, id, parent id, timestamp, cwd, type, and payload;
- supports message, model change, thinking change, label, compaction, branch
  summary, and extension entries.
- preserves unknown top-level fields across marshal/unmarshal.
- follows the envelope fields in `spec/protocols/session-jsonl.md`.

## Functions

### `MarshalEntry(e Entry) ([]byte, error)`

Logic:

- Validate required envelope fields: schema, id, kind, created_at.
- Validate `parent_id` shape but not existence.
- Marshal known payload type according to `kind`.
- Preserve `extra` fields after known fields without allowing them to override
  required fields.
- Return bytes without trailing newline; caller owns JSONL delimiter.

Acceptance:

- produces one JSON object without trailing newline.

### `UnmarshalEntry(line []byte) (Entry, error)`

Logic:

- Trim one trailing CR when reading CRLF JSONL.
- Decode envelope into raw payload first.
- Reject missing schema/id/kind/created_at.
- Store unknown optional top-level fields as `extra`.
- Return typed error with line context supplied by caller/store.

Acceptance:

- rejects unknown required fields and malformed parents later in store.

Tests:

- `TestNUF080SessionAppendBuildsTree`
- `TestSessionEntryPreservesExtraFields`
