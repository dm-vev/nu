# `internal/session/store.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 2931429
Implementation Comments: Append/load use direct stdlib filesystem calls with in-process locking; Phase 4 adds lookup/fork/clone/import/export helpers.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Append-only JSONL session storage.
Implements file rules from `spec/protocols/session-jsonl.md`.

## Code Style

Filesystem writes are atomic enough for single-process use. Use explicit locks
for concurrent append in the same process.

## Functions

### `type Ref struct`

Logic:

- Identify a session by id.
- Optionally carry an explicit JSONL path for path-based resume.
- Prefer explicit path when present; otherwise resolve under store root.

Acceptance:

- path resume can load a session outside the root without changing store root.

### `OpenStore(root string) *Store`

Logic:

- Clean the root path.
- Initialize no session files and start no background work.
- Return a concrete store safe for temp-dir tests.

Acceptance:

- stores no global paths;
- can be rooted in a temp directory in tests.

### `(*Store) Append(ctx context.Context, ref Ref, entry Entry) error`

Logic:

- Validate session ref, entry id, schema, kind, and parent id shape.
- Acquire the in-process lock before parent validation and append.
- Create parent directories.
- Open file append-only; create header first when creating a new session.
- Marshal entry and append LF.
- Release locks with defer and wrap path-qualified errors.

Acceptance:

- appends one JSONL line;
- rejects entry with missing id.

### `(*Store) Resolve(ctx context.Context, selector string) (Ref, error)`

Logic:

- Treat an existing path or `.jsonl` selector as a direct session path.
- Decode the header to fill the returned ref id.
- Otherwise match selector as a full or partial session id under the store root.
- Return not-found or ambiguous errors for zero/multiple matches.

Acceptance:

- resumes by explicit JSONL path;
- resumes by unambiguous partial id.

### `(*Store) LatestByCWD(ctx context.Context, cwd string) (Ref, error)`

Logic:

- Scan session files under the store root.
- Decode headers only.
- Match cleaned `cwd` against header cwd.
- Pick the newest matching header by `created_at`, breaking ties by id.

Acceptance:

- `--continue` can find the latest session for the current working directory.

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

### `(*Store) Fork(ctx context.Context, source Ref, target Ref, entryID string) error`

Logic:

- Load the source session.
- Build the root-to-entry path for `entryID`.
- Create a new target session containing exactly that path.
- Reject an already existing target file.

Acceptance:

- forked session starts at the selected entry and keeps parent links valid.

### `(*Store) Clone(ctx context.Context, source Ref, target Ref) error`

Logic:

- Load the source session.
- Build the active branch path.
- Create a new target session containing that active branch.
- Reject an already existing target file.

Acceptance:

- cloned session contains only the active branch.

### `(*Store) Export(ctx context.Context, ref Ref, w io.Writer) error`

Logic:

- Load the session first to validate it.
- Copy the original JSONL bytes to `w`.
- Keep the output byte-for-byte with the stored file.

Acceptance:

- exported sessions can be re-imported.

### `(*Store) Import(ctx context.Context, r io.Reader, target Ref) (Ref, error)`

Logic:

- Read the JSONL payload.
- Decode and validate header and entries.
- Build the tree before writing.
- Use `target.ID` when supplied, otherwise keep the imported header id.
- Reject an already existing target file.

Acceptance:

- invalid imports do not create files;
- valid imports round-trip through `Load`.

Tests:

- `TestNUF080SessionAppendBuildsTree`
- `TestNUF080SessionLoadRejectsBrokenParent`
- `TestNUF080SessionAppendRejectsBrokenParent`
- `TestSessionAppendRejectsDuplicateID`
- `TestSessionExportImportRoundTrip`
- `TestNUF081ContinueLatestByCWD`
- `TestNUF081ResumeByPathOrPartialID`
- `TestNUF081ForkStartsNewFileFromUserEntry`
- `TestNUF081CloneCopiesActiveBranch`
