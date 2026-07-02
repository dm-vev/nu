# `internal/session/store.go`

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

Append-only JSONL session storage.
Implements file rules from `spec/protocols/session-jsonl.md`.

## Code Style

Filesystem writes are atomic enough for single-process use. Use explicit locks
for concurrent append in the same process.

## Functions

### `OpenStore(root string, fs FS) *Store`

Logic:

- Clean the root path and store the injected filesystem/process options.
- Initialize no session files and start no background work.
- Set default lock, clock, id, and fsync options when not provided.
- Return a concrete store safe for temp-home tests.

Acceptance:

- stores no global paths;
- can be backed by temp filesystem in tests.

### `(*Store) Append(ctx context.Context, ref Ref, entry Entry) error`

Logic:

- Validate session ref, entry id, schema, kind, and parent id shape.
- Acquire per-file in-process lock and optional advisory file lock.
- Create parent directories.
- Open file append-only; create header first when creating a new session.
- Marshal entry, append LF, optionally fsync.
- Release locks with defer and wrap path-qualified errors.

Acceptance:

- appends one JSONL line;
- fsync behavior is configurable for tests/release;
- rejects entry with missing id.

### `(*Store) Load(ctx context.Context, ref Ref) (*Session, error)`

Logic:

- Resolve ref to a concrete session file.
- Read line by line using LF framing.
- Decode and validate header before entries.
- Unmarshal entries with line numbers.
- Pass entries to `BuildTree`.
- Determine active branch by `spec/protocols/session-jsonl.md` rules.
- Return loaded session plus diagnostics for optional payload issues.

Acceptance:

- reconstructs tree;
- rejects broken parent links;
- returns active branch.

Tests:

- `TestNUF080SessionAppendBuildsTree`
- `TestNUF080SessionLoadRejectsBrokenParent`
