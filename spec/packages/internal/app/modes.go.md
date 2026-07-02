# `internal/app/modes.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Help/version dispatch exists. Print mode now calls the runtime agent directly.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

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
- Return a clear not-implemented error for modes outside the current slice.

Acceptance:

- dispatches help, version, and print mode;
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

Tests:

- `TestNUF002DispatchPrintMode`
- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestAppRunPrintModeWithoutHandlerFails`
