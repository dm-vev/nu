# `internal/app/modes.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 687e919
Implementation Comments: Help/version/print/list-models/JSON dispatch remains intact. RPC mode creates an RPC server first and injects an agent emitter. Interactive mode creates a TUI app first and injects an agent emitter. List-models includes optional display names.

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
- Dispatch list-models to `runListModels`.
- Dispatch chat print mode to `runPrint`.
- Dispatch chat JSON mode to `runJSON`.
- Dispatch chat RPC mode to `runRPC`.
- Dispatch chat interactive mode to `runInteractive`.
- Configure provider/runtime selections from the parsed request before chat
  modes create an agent.
- Return a clear not-implemented error for modes outside the current slice.

Acceptance:

- dispatches help, version, list-models, print mode, and JSON mode;
- dispatches RPC and interactive modes;
- configured real providers are reachable from print and JSON modes;
- returns a clear error for unimplemented modes.

### `runListModels(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Load auth from `Options.Home` and `Options.Env`.
- Resolve provider auth state through runtime helpers.
- Build the model registry from built-ins plus optional `req.ModelsPath`.
- Print visible models as tab-separated `provider/id`, api, context window,
  max output, and optional display name fields.
- Keep output deterministic by using registry ordering.

Acceptance:

- authenticated provider models are visible;
- unauthenticated provider models are hidden;
- custom `--models` entries are included.
- custom `display_name` entries are visible in list output.

### `runPrint(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Require a provider-backed agent.
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

### `runRPC(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Create an `rpc.Server` over runtime stdin/stdout/stderr and runtime labels.
- Create a provider-backed agent with `server.Emit`.
- Inject the agent into the server.
- Run the server until stdin EOF, shutdown command, context cancellation, or
  protocol write error.

Acceptance:

- `--mode rpc` accepts JSONL commands from stdin;
- stdout contains only JSONL responses/events.

### `runInteractive(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Create a `tui.App` over runtime stdin/stdout/stderr and runtime labels.
- Create a provider-backed agent with `ui.Emit`.
- Inject the agent into the UI app.
- Run the UI loop until quit input, stdin EOF, context cancellation, or UI write
  error.

Acceptance:

- `--mode interactive` starts a deterministic UI loop;
- non-empty submitted lines call the provider-backed agent.

Tests:

- `TestNUF002DispatchPrintMode`
- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestAppRunPrintModeWithoutAvailableModelFails`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
- `TestNUF170JSONModeFeedsToolResultBackToProvider`
- `TestListModelsUsesAuthState`
- `TestListModelsUsesCustomModelsPath`
- `TestListModelsIncludesDisplayName`
- `TestNUF002DispatchRPCMode`
- `TestNUF002DispatchInteractiveMode`
