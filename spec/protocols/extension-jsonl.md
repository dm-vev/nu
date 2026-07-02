# Extension JSONL Protocol

## Purpose

Run extensions out of process without making Node, Go plugins, or any one
language runtime part of the core binary.

## Process Model

- Nu starts an extension process with stdin/stdout reserved for JSONL protocol.
- Extension stderr is diagnostic text and may be shown/logged.
- First extension frame must be `hello`.
- Nu replies with `hello_ack` or terminates the process.
- Extension process is session-scoped unless marked global-safe by manifest.

## Framing

- One JSON object per LF-delimited line.
- Every request frame has `id`.
- Responses repeat `id`.
- Events without responses omit `id`.

## Handshake

Extension sends:

```json
{
  "type": "hello",
  "protocol": 1,
  "extension_id": "name-or-path-hash",
  "capabilities": ["tools", "commands", "hooks", "ui"]
}
```

Nu replies:

```json
{
  "type": "hello_ack",
  "protocol": 1,
  "session_id": "...",
  "mode": "interactive|print|json|rpc"
}
```

## Registration Frames

Extensions may register:

- tool definitions;
- slash commands;
- keybinding actions;
- CLI flags;
- provider/model entries;
- renderers;
- lifecycle hooks.

Registration must complete before `session_start` unless a later command
explicitly allows dynamic registration.

## Hook Frames

Nu sends hook requests in load order. Hook response:

```json
{
  "id": "request-id",
  "type": "hook_response",
  "action": "continue|modify|block",
  "reason": "optional",
  "payload": {}
}
```

Rules:

- `block` short-circuits tool/session operations where blocking is allowed;
- `modify` must validate against the target payload schema;
- hook timeout is configurable and becomes a hook error.

## UI Frames

Supported UI requests:

- `notify`
- `select`
- `confirm`
- `input`
- `status`
- `widget`
- `custom`

Headless modes reject interactive requests unless the request provides a default
value.

## Shutdown

Nu sends:

```json
{ "id": "...", "type": "shutdown", "reason": "normal|error|reload" }
```

Extension must respond before timeout. Nu kills the process after timeout.
Shutdown is sent at most once per process.

## Security Rules

- Project extensions load only after trust.
- Extension paths and package sources appear in diagnostics and session metadata.
- Extension tools are disabled by `--no-tools`.
- Extension process inherits only the environment Nu decides to pass.

## Tests

- handshake version mismatch rejects extension;
- tool registration appears in registry;
- hook blocks dangerous tool call;
- interactive UI request rejected in print mode without default;
- shutdown runs once and process is killed on timeout.
