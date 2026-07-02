# `internal/export/jsonl.go`

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

Export and import Nu session JSONL.

## Code Style

Thin wrapper over session marshaling with user-facing diagnostics.

## Functions

### `WriteJSONL(ctx context.Context, w io.Writer, sess *session.Session) error`

Logic:

- Check context between entries for long exports.
- Iterate persisted entries in append order, not active-branch order.
- Marshal each entry with `session.MarshalEntry` and add exactly one LF delimiter.
- Return writer errors with the entry id being exported.

Acceptance:

- writes entries in append order;
- preserves all payload details.

### `ReadJSONL(ctx context.Context, r io.Reader) (*session.Session, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Round-trips exported sessions.

Acceptance:

- round-trips exported sessions.

Tests:

- `TestNUF180ExportJSONLRoundTrip`
