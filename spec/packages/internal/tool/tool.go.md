# `internal/tool/tool.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 6ec7970
Implementation Comments: Root tool package aggregates one-tool subpackages into the agent tool map and exposes provider-facing tool schemas.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Expose the Phase 2 built-in tool set to `internal/app` without putting tool
implementations in the root package.

## Code Style

Keep this file thin. It imports each one-tool subpackage and wires function
closures only. No filesystem, process, search, or mutation logic belongs here.

## Functions

### `Builtins(cwd string) map[string]agent.ToolFunc`

Logic:

- Return exactly seven entries: `read`, `write`, `edit`, `bash`, `grep`,
  `find`, and `ls`.
- For output-producing tools, pass the shared default max output bytes.
- For mutation tools, delegate to their subpackage without extra wrapping.
- Start no background work and touch no filesystem during construction.

Acceptance:

- app can pass returned tools directly into `agent.Options`;
- built-in JSON mode can execute `read` without `Options.Tools`.

Tests:

- `TestBuiltinsExposesEveryPhaseTwoTool`
- `TestJSONModeUsesBuiltinToolsByDefault`

### `Definitions() []provider.ToolDefinition`

Logic:

- Return provider-facing definitions for every built-in tool.
- Include short descriptions and JSON object parameter schemas.
- Keep definitions side-effect free.

Acceptance:

- OpenAI-compatible providers receive callable tool schemas.
- `bash` is advertised with a required `command` parameter.

Tests:

- `TestDefinitionsExposeBashSchema`

### `objectSchema(properties map[string]any, required []string) map[string]any`

Logic:

- Build a minimal JSON object schema with `additionalProperties: false`.
- Normalize nil `required` to an empty array.

Acceptance:

- Tool definitions are valid JSON schema objects.
