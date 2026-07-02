# `internal/tool/toolkit/toolkit.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 3d3fb26
Implementation Comments: Shared stdlib helpers support tool subpackages; path sandboxing resolves symlinks before accepting a cwd-contained path.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Avoid duplicating JSON argument decoding, cwd path sandboxing, JSON result
formatting, truncation, glob/gitignore matching, and the Phase 2 mutation lock.

## Code Style

Small helper functions only. No subpackage imports except stdlib and
`internal/agent`. Helpers must be boring and deterministic.

## Functions

### `DecodeArgs(raw string, out any) error`

Logic:

- Treat empty raw JSON as `{}`.
- Decode into `out`.
- Wrap JSON errors with `decode tool args`.

Acceptance:

- supports empty argument objects;
- returns contextual JSON errors.

### `ResolveUnder(cwd, requested string) (path string, rel string, err error)`

Logic:

- Require non-empty cwd and relative requested path.
- Clean requested path.
- Reject `..` escapes and absolute paths.
- Resolve cwd and the requested path through symlinks.
- For create targets that do not exist yet, resolve the nearest existing parent
  and join missing leaf components only after containment is proven.
- Reject resolved paths outside cwd.
- Return resolved absolute path under cwd and slash-form relative path.

Acceptance:

- rejects lexical cwd escapes;
- rejects symlink cwd escapes for existing files and new files under symlinked
  parents;
- supports `.` as root.

### `JSONResult(value map[string]any) (agent.ToolResult, error)`

Logic:

- Marshal one JSON object.
- Return it as `agent.ToolResult.Content`.
- Wrap marshal errors.

Acceptance:

- tool results are JSON strings.

### `JSONListResult(key string, values []string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Marshal `{key: values, truncated: bool}`.
- Use an empty JSON array, not `null`, for empty lists.
- Remove trailing values until the JSON result fits `maxBytes`.

Acceptance:

- preserves deterministic list order;
- marks truncation when values are dropped.

### Other Helpers

Logic:

- `Relative` returns slash-form cwd-relative paths.
- `TruncateString` cuts display output by byte budget.
- `ImageMime` recognizes png, jpg/jpeg, gif, and webp by extension.
- `PersistTempOutput` writes full truncated command output to a temp file.
- `LoadGitignore`, `ShouldSkip`, and `GlobMatches` implement Phase 2 simple
  root `.gitignore` and glob behavior.
- `MutationMu` serializes write/edit mutations in Phase 2.

Acceptance:

- helpers are covered through subpackage tests.

Tests:

- covered by read/write/edit/bash/grep/find/ls tests.
