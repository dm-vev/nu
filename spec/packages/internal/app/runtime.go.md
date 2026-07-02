# `internal/app/runtime.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 0f96e6e
Implementation Comments: Runtime carries process IO, provider settings, tools, optional session id, mode-specific emitters, and default built-ins when no test tools are supplied.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Shared runtime object passed to mode handlers.

## Code Style

Plain structs with concrete dependencies. Construct mode-specific agents with
the emitter that mode needs; do not add registries before there are multiple
adapters.

## Types

### `type Runtime struct`

Logic:

- Store normalized process IO, provider settings, tools, and session id.
- Keep ownership explicit: runtime closes only components it created.
- Do not start goroutines, TUI loops, RPC loops, or provider streams during construction.

Acceptance:

- contains normalized process IO, provider settings, tools, and session id;
- has no goroutines after construction.

### `type Options struct`

Logic:

- Carry argv, environment, cwd, home, stdin, stdout, stderr, version, optional
  provider settings, tool functions, and optional session id from `cmd/nu`.
- Keep options serializable enough for integration tests to construct without process globals.
- Do not store parsed CLI state here; parsing output belongs to `cli.Request`.

Acceptance:

- carries argv, env, cwd, home, stdin, stdout, stderr, version metadata,
  optional provider, tool functions, and optional session id.

## Functions

### `newAgent(opts Options, emit func(agent.Event)) *agent.Agent`

Logic:

- Return nil when no provider is configured.
- Create `agent.Agent` with provider id, API, model, and provider stream.
- Pass configured tool functions through to the agent.
- Use `tool.Builtins(opts.CWD)` when no tool map is supplied.
- Install the mode-specific event callback.

Acceptance:

- no provider means no agent;
- mode-specific emitter receives agent events.
- default built-ins are available when `Options.Tools` is nil.

### `newJSONSessionHeader(opts Options) (jsonSessionHeader, error)`

Logic:

- Use `opts.SessionID` when supplied.
- Generate a UUIDv4-like id when no session id is supplied.
- Default empty app version to `dev`.
- Fill the session header fields used by JSON mode.

Acceptance:

- produces a session header with id, schema, cwd, app, and app version.

### `writeJSONLine(w io.Writer, value any) error`

Logic:

- Marshal `value` with `encoding/json`.
- Write bytes plus one LF.
- Wrap marshal/write errors with context.

Acceptance:

- writes exactly one valid JSON object line per call.

Tests:

- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
- `TestJSONSessionHeaderDefaults`
- `TestJSONModeUsesBuiltinToolsByDefault`
