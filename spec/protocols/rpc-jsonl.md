# RPC JSONL Protocol

## Purpose

Allow editors and headless clients to drive Nu over stdin/stdout.

## Framing

- One JSON object per LF-delimited line.
- Readers split only on `\n`.
- A single trailing `\r` before `\n` is ignored.
- Stdout is protocol-only.
- Stderr is diagnostics-only.

## Common Fields

Every client command:

```json
{ "id": "optional-client-id", "type": "prompt" }
```

Every command response:

```json
{
  "id": "same-id-if-provided",
  "type": "response",
  "command": "prompt",
  "success": true,
  "error": null,
  "data": {}
}
```

Events are the same event objects as JSON mode, plus optional RPC envelope only
when needed for correlation.

## Commands

### `prompt`

Fields:

- `message`
- `images`, optional
- `streaming_behavior`: absent, `steer`, or `follow_up`

Rules:

- when agent idle, accepted immediately;
- when agent active, must specify streaming behavior unless the message is an
  immediate extension command.

### `steer`

Fields:

- `message`
- `images`, optional

Rules:

- queues steering for next provider call;
- rejects extension commands.

### `follow_up`

Fields:

- `message`
- `images`, optional

Rules:

- queues until agent becomes idle;
- rejects extension commands.

### `abort`

Rules:

- cancels active provider/tool work;
- response means abort request was accepted, not that all work has fully
  unwound.

### `new_session`

Fields:

- `parent_session`, optional

Rules:

- creates new session after extension hooks allow it.

### `state`

Rules:

- returns current session id, cwd, model, queues, busy state, and active leaf.

### `set_settings`

Fields:

- `settings`: partial settings object

Rules:

- applies runtime-supported settings immediately;
- persists only when `persist` is true.

### `shutdown`

Rules:

- stops RPC server after pending response is written.

## Error Rules

- Invalid JSON returns a response only when an id can be extracted; otherwise an
  uncorrelated error event is emitted.
- Accepted prompts report later provider/tool failures through event stream, not
  a second response.

## Tests

- response id correlation;
- stdout contains only JSONL;
- prompt during active stream rejected without behavior;
- steer delivered before next provider call;
- shutdown writes final response.
