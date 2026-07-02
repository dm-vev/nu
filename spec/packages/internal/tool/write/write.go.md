# `internal/tool/write/write.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 6ec7970
Implementation Comments: Write tool lives in its own subpackage and uses the shared Phase 2 mutation lock.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement the `write` built-in tool.

## Code Style

Use stdlib filesystem calls. Use the Phase 2 shared mutation lock; split to
per-path locks only when throughput requires it.

## Functions

### `Run(ctx context.Context, cwd string, raw string) (agent.ToolResult, error)`

Logic:

- Decode `path` and `content`.
- Resolve path under cwd and reject escapes.
- Check cancellation before mutation.
- Acquire `toolkit.MutationMu`.
- Create parent directories.
- Write content with `0644` permissions.
- Return JSON with `path` and byte count.

Acceptance:

- creates nested files;
- overwrites existing files;
- concurrent same-path writes are serialized enough to avoid partial content;
- rejects cwd escapes.

Tests:

- `TestNUF071WriteCreatesFile`
- `TestNUF071ConcurrentWritesSamePathSerialize`
- `TestWriteRejectsPathEscape`
