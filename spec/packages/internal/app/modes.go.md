# `internal/app/modes.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 456582c
Implementation Comments: Help/version/print dispatch exists. JSON mode writes JSONL session header and agent events.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [x] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Dispatch resolved CLI requests to the modes implemented in the current slice.

## Code Style

Use a small switch over command kind. Mode handlers return `error`; conversion
to exit code stays in `app.go`.

## Functions

### `runMode(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Dispatch help and version without constructing extra state.
- Dispatch chat print mode to `runPrint`.
- Dispatch chat JSON mode to `runJSON`.
- Return a clear not-implemented error for modes outside the current slice.

Acceptance:

- dispatches help, version, print mode, and JSON mode;
- returns a clear error for unimplemented modes.

### `runPrint(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Require `rt.Agent`; do not accept an injected print callback.
- Join prompt args into the prompt text exactly as CLI parsing supplied them.
- Call `Agent.Prompt` with the caller context.
- Let the runtime agent emitter own stdout writing.

Acceptance:

- calls the agent provider path for `--print`;
- fails clearly when no provider-backed agent exists.

### `runJSON(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Require a provider-backed agent.
- Write one session header JSON object before agent events.
- Write every agent event as one JSON object per line.
- Keep human diagnostics out of stdout by returning errors to `app.Run`.

Acceptance:

- stdout is valid JSONL only;
- first line is the session header;
- later lines are agent events in emission order.

Tests:

- `TestNUF002DispatchPrintMode`
- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestAppRunPrintModeWithoutHandlerFails`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
- `TestNUF170JSONModeFeedsToolResultBackToProvider`
