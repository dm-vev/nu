# `internal/rpc/server.go`

## Status

Current: IMPLEMENTED
Implementation Commit: 5d9629b
Implementation Comments: Server recognizes the Pi RPC command set, forwards agent events as JSONL, supports prompt busy rejection, steering/follow-up queues, state/model/settings mutation, built-in bash execution, EOF shutdown, and in-memory session/tree responses until the durable session controller is wired.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Implement Pi-compatible headless JSONL command wiring over an injected Nu
agent, stdout writer, and in-memory runtime/session state.

## Code Style

Keep RPC as a thin orchestration layer. It may import `agent` and built-in tool
helpers, but it must not import `app`. Use one mutex for protocol writes and
one mutex for mutable runtime state. Comments are required around asynchronous
prompt start/finish and queue draining.

## Types

### `type Options struct`

Logic:

- Carry stdin, stdout, stderr, cwd, session id/name, provider/api/model labels,
  and the optional agent pointer.
- Keep all fields injectable for tests.

Acceptance:

- tests can run the server with fake IO and fake providers;
- no process globals are required.

### `type Server struct`

Logic:

- Own protocol IO, mutable state, prompt/follow-up queues, runtime settings,
  in-memory messages, and the optional agent.
- Expose `Emit` so the app can build an agent whose events are forwarded to RPC
  stdout.
- Expose `SetAgent` so construction does not require cyclic dependencies.

Acceptance:

- agent events and command responses share stdout safely;
- server state can be queried without a provider call.

## Functions

### `NewServer(opts Options) *Server`

Logic:

- Normalize nil IO to empty reader/discard writers.
- Seed session id, cwd, provider, api, model, and default queue/settings values.
- Return an idle server without starting goroutines.

Acceptance:

- construction has no provider side effects.

### `(*Server) SetAgent(a *agent.Agent)`

Logic:

- Store the agent under the state mutex.
- Allow nil only for state-only tests.

Acceptance:

- app can create server first, then pass `server.Emit` to `agent.New`.

### `(*Server) Emit(ev agent.Event)`

Logic:

- Update in-memory messages for durable event types.
- Write the event as JSONL to stdout.
- Store the first write error for `Run` to return.

Acceptance:

- provider/tool events pass through unchanged;
- stdout remains valid JSONL.

### `(*Server) Run(ctx context.Context) error`

Logic:

- Read stdin as strict JSONL.
- Parse each line into a generic command.
- Route extension UI responses without treating them as unknown commands.
- For each command, write exactly one response unless the command intentionally
  starts async work that writes its own response.
- Stop after `shutdown`, EOF, context cancellation, or a stored write error.

Acceptance:

- response ids are preserved;
- invalid JSON produces a structured error response;
- `shutdown` writes its response before returning.

### `(*Server) handleCommand(ctx context.Context, command commandEnvelope) (response, bool)`

Logic:

- Dispatch every Pi RPC command listed in `NUF-171`.
- For `prompt`, reject when busy unless streaming behavior is `steer` or
  `follow_up`; otherwise start prompt work asynchronously and respond once
  accepted.
- For queue commands, append to the relevant queue and emit queue state events.
- For state/session/model/settings commands, mutate or report in-memory state.
- For bash, call the built-in bash tool with the server cwd.
- For unsupported subsystem depth, return a clear structured error instead of
  silently ignoring the command.

Acceptance:

- Pi command names are recognized;
- prompt busy rejection matches the protocol;
- headless clients can inspect state after every command.

### `(*Server) startPrompt(ctx context.Context, id string, text string) response`

Logic:

- Require an agent.
- Mark the server busy before launching the goroutine so immediate follow-up
  commands observe the active stream.
- Merge queued steering text into the prompt before the provider call.
- Record the user message before provider work starts.
- Run `agent.Prompt` in a goroutine.
- On finish, clear busy state, emit a prompt error event when needed, and drain
  one queued follow-up if present.

Acceptance:

- prompt response is immediate and correlated;
- steering is delivered before the next provider request;
- follow-up prompts run after the current prompt becomes idle.

Tests:

- `TestNUF171RPCPromptResponseCorrelation`
- `TestNUF171RPCRejectsPromptDuringStreamWithoutBehavior`
- `TestNUF171RPCShutdownWritesFinalResponse`
- `TestNUF052SteeringDeliveredBeforeNextProviderCall`
- `TestNUF171RPCRecognizesPiCommandSet`
