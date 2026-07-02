# `internal/tool/builtin/read.go`

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

Built-in `read` tool.

## Code Style

No terminal rendering here. Return structured content and details.

## Functions

### `NewRead(cwd string, opts ReadOptions) tool.Tool`

Logic:

- Copy constructor inputs into a concrete value without starting background work.
- Apply defaults before returning the constructed value.
- Registers name `read`, schema, and description.

Acceptance:

- registers name `read`, schema, and description.

### `executeRead(ctx context.Context, args ReadArgs, ops ReadOps) tool.Result`

Logic:

- Resolve the requested path using read-path compatibility rules.
- Validate offset, byte/line limits, binary policy, and image policy before opening the file.
- For text files, read the requested range and apply truncation metadata.
- For supported images, return an image content block unless image reads are disabled.
- Return not-found and unsupported-file cases as tool errors, not Go errors.

Acceptance:

- reads text with offset/limit;
- truncates large output;
- returns image attachment for supported images unless blocked;
- reports file not found as tool error.

Tests:

- `TestNUF070ReadTextWithOffsetLimit`
- `TestNUF070ReadTruncatesLargeFile`
- `TestNUF070ReadImageAttachment`
