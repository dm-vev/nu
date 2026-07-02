# `internal/export/html.go`

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

Export standalone HTML session view.

## Code Style

Use `html/template`. Escape all user/model/tool content.

## Functions

### `WriteHTML(ctx context.Context, w io.Writer, sess *session.Session, opts HTMLOptions) error`

Logic:

- Check context before rendering large sessions.
- Walk active session entries in display order and convert each known payload to escaped HTML.
- Render messages, thinking, tool calls, tool outputs, images, metadata, branch summaries, and compaction summaries.
- Embed only the CSS/JS required for standalone viewing and never write unescaped session content.

Acceptance:

- renders messages, thinking, tool calls, tool outputs, images, and metadata;
- embeds needed CSS/JS;
- escapes untrusted content.

Tests:

- `TestNUF180ExportHTMLContainsEscapedContent`
