# Session JSONL Protocol

## Purpose

Persist conversations as append-only branchable trees. Nu has its own versioned
format and imports Pi sessions explicitly.

## File Rules

- Encoding: UTF-8 JSON Lines.
- Delimiter: LF. Readers may accept CRLF by trimming one trailing CR.
- One JSON object per line.
- Existing lines are never rewritten by normal operation.
- Appending a line is the only mutation except explicit maintenance/import
  commands.

## Header

First line:

```json
{
  "type": "session",
  "schema": 1,
  "id": "uuid",
  "created_at": "2026-07-02T00:00:00Z",
  "cwd": "/abs/path",
  "app": "nu",
  "app_version": "0.0.0"
}
```

Rules:

- `schema` is required.
- `id` is stable for the file.
- `cwd` is absolute at creation time.

## Entry Shape

Every non-header line:

```json
{
  "type": "entry",
  "schema": 1,
  "id": "uuid",
  "parent_id": "uuid-or-empty",
  "created_at": "2026-07-02T00:00:00Z",
  "kind": "message",
  "payload": {}
}
```

Rules:

- `id` is unique within the file.
- `parent_id` is empty only for root entries.
- `parent_id` must reference a previous entry, never a later entry.
- `kind` decides payload schema.
- unknown optional fields are preserved in `extra` when imported.

## Entry Kinds

- `message`
- `model_change`
- `thinking_change`
- `label`
- `compaction`
- `branch_summary`
- `extension`

## Message Payload

```json
{
  "role": "user|assistant|tool_result|bash_execution|custom|branch_summary|compaction_summary",
  "timestamp_ms": 1780000000000,
  "content": []
}
```

Assistant payload includes `provider`, `api`, `model`, `usage`, `stop_reason`,
and optional `error_message`.

Tool result payload includes `tool_call_id`, `tool_name`, `is_error`,
`content`, and optional `details`.

## Content Blocks

- `text`: `{ "type": "text", "text": "..." }`
- `image`: `{ "type": "image", "mime_type": "image/png", "data": "base64" }`
- `thinking`: `{ "type": "thinking", "thinking": "..." }`
- `tool_call`: `{ "type": "tool_call", "id": "...", "name": "...", "arguments": {} }`

## Active Branch

The active branch is not a mutable pointer in the file. On load:

1. Build tree from all entries.
2. If a final `session_state` extension entry exists and references a valid
   leaf, use it.
3. Otherwise use the last appended entry as active leaf.

If future UX needs explicit active-leaf persistence, add a new entry kind rather
than rewriting the header.

## Validation

Load must reject:

- missing header;
- duplicate entry id;
- parent id that points nowhere;
- parent id that points to a later line;
- unknown required schema version;
- malformed JSON line.

Load may keep diagnostics and continue only for explicitly optional payload
fields.

## Append Locking

Within one process, append is serialized per session file. Cross-process locking
uses an advisory lock file next to the session file when platform support is
available. If locking fails, append returns an error unless caller explicitly
opens read-only mode.

## Tests

- header round-trip;
- append builds tree;
- duplicate id rejection;
- parent-to-future-line rejection;
- active branch fallback to last appended entry;
- Pi import preserves message/tool content.
