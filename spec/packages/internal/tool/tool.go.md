# `internal/tool/tool.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 0f96e6e
Implementation Comments: Phase 2 built-ins live in one stdlib-only file; write/edit use one global mutation lock until throughput proves per-path locks are needed.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Provide the built-in coding tools used by the agent in Phase 2: read, write,
edit, bash, grep, find, and ls.

## Code Style

Use the standard library only. Keep arguments as JSON decoded inside each tool.
Resolve paths under cwd and reject paths escaping cwd. Keep truncation simple and
deterministic.

## Functions

### `Builtins(cwd string) map[string]agent.ToolFunc`

Logic:

- Return a map for read, write, edit, bash, grep, find, and ls.
- Each function decodes raw JSON arguments from `agent.ToolCall.Arguments`.
- Use default output truncation for provider-facing output.
- Do not start background work during construction.

Acceptance:

- app can pass returned tools directly into `agent.Options`.

### `Read(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode path, optional offset, and optional limit.
- Resolve path under cwd and reject directories.
- For image extensions, return mime type and base64 data.
- For text, apply byte offset/limit, then truncate output to maxBytes.

Acceptance:

- reads text with offset/limit;
- truncates large text;
- reads supported images as base64 attachments.

### `Write(ctx context.Context, cwd string, raw string) (agent.ToolResult, error)`

Logic:

- Decode path and content.
- Resolve path under cwd.
- Create parent directories.
- Write content with user-readable file permissions.

Acceptance:

- creates or overwrites a file under cwd;
- rejects paths escaping cwd.

### `Edit(ctx context.Context, cwd string, raw string) (agent.ToolResult, error)`

Logic:

- Decode path and replacement list.
- Read original content.
- For every replacement, require exactly one match in the original content.
- Apply replacements and write the result.
- Return a small unified patch.

Acceptance:

- applies exact replacement;
- rejects missing or ambiguous old text;
- preserves existing line endings unless replacement changes them.

### `Bash(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode command and optional timeout milliseconds.
- Run through `sh -c` in cwd with context cancellation.
- Capture stdout, stderr, exit code, and timeout state.
- Truncate displayed output to maxBytes and write full output to a temp file
  when truncation happens.

Acceptance:

- captures stdout/stderr/exit code;
- kills timed-out commands;
- persists full output when display output is truncated.

### `Grep(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode pattern, optional literal, ignore-case, glob, root, and limit.
- Walk files under root, skipping `.git` and root `.gitignore` matches.
- Match by literal substring or regexp.
- Return sorted `path:line:text` matches with truncation.

Acceptance:

- supports literal and regexp matching;
- respects root `.gitignore`.

### `Find(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode root, glob, and limit.
- Walk files under root, skipping `.git` and root `.gitignore` matches.
- Return sorted relative paths.

Acceptance:

- finds files by glob;
- respects root `.gitignore`;
- enforces limit.

### `Ls(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error)`

Logic:

- Decode optional path.
- Resolve directory under cwd.
- List entries sorted alphabetically.
- Include dotfiles and append `/` to directories.
- Truncate output to maxBytes.

Acceptance:

- lists sorted entries with directories marked;
- rejects non-directories.

Tests:

- `TestNUF070ReadTextWithOffsetLimit`
- `TestNUF070ReadTruncatesLargeFile`
- `TestNUF070ReadImageAttachment`
- `TestNUF071WriteCreatesFile`
- `TestNUF071ConcurrentWritesSamePathSerialize`
- `TestNUF072EditSingleReplacement`
- `TestNUF072EditRejectsAmbiguousOldText`
- `TestNUF072EditPreservesCRLF`
- `TestNUF073BashCapturesStdoutAndStderr`
- `TestNUF073BashTimeoutKillsProcess`
- `TestNUF073BashTruncatesAndPersistsFullOutput`
- `TestNUF074GrepLiteralAndRegex`
- `TestNUF074GrepRespectsGitignore`
- `TestNUF075FindGlob`
- `TestNUF075FindRespectsGitignore`
- `TestNUF076LsSortedWithDirs`
- `TestNUF076LsRejectsNonDirectory`
