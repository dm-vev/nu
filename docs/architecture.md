# Nu Architecture

Nu is built as a Go application with a small executable entry point and private
packages under `internal/`.

## Layers

```text
cmd/nu
  -> internal/app
     -> cli/config/auth/resource/model
     -> agent/session/tool/provider/compact
     -> tui or json/rpc/export/share/update
```

The application layer wires dependencies. Lower packages do not import `app`.

## Core Interfaces

These are conceptual shapes; exact Go names are decided when tests are written.

```go
type Provider interface {
    Stream(ctx context.Context, req Request) (<-chan Event, error)
}

type Tool interface {
    Name() string
    Schema() Schema
    Execute(ctx context.Context, call ToolCall) ToolResult
}

type Store interface {
    Append(ctx context.Context, entry Entry) error
    Load(ctx context.Context, id SessionRef) (*Session, error)
}
```

Interfaces stay where they are consumed. Provider packages return concrete
clients.

## Protocol Contracts

The stable contracts live under `spec/protocols/`:

- provider streams: normalized event ordering and tool-call assembly;
- session JSONL: append-only tree storage;
- RPC JSONL: headless stdin/stdout framing;
- extension JSONL: process extension handshake, registration, hooks, UI, and
  shutdown;
- TUI rendering/input: frame width, diff, input, editor, and overlay invariants.

Package specs can change as implementation teaches us more. Protocol specs need
golden-test updates before any wire or persisted format changes.

## Data Flow

### Print Mode

```text
args -> settings/auth/model -> session -> agent -> provider/tools -> stdout
```

Print mode is the first full vertical slice because it exercises model calls,
tools, events, sessions, and output without terminal complexity.

### Interactive Mode

```text
terminal input -> tui editor -> slash/agent command -> agent events -> tui render
```

The TUI is an event consumer. It cannot mutate agent state directly; it sends
commands to the session controller.

### RPC Mode

```text
stdin JSONL command -> rpc dispatcher -> session controller -> stdout JSONL
```

RPC stdout must remain machine-readable JSONL. Diagnostics go to stderr.

## Package Responsibilities

### `internal/agent`

Owns turn state, provider streaming, tool call scheduling, retry, abort,
steering, follow-up queues, and event emission.

### `internal/session`

Owns JSONL storage, tree reconstruction, active branch selection, resume lookup,
fork, clone, import, and migration.

### `internal/provider`

Defines shared request/message/event structs used by adapters. Adapters own
their HTTP payload shape and stream parsing.

### `internal/tool`

Owns registry and validation. Built-ins are boring Go code over filesystem,
subprocesses, and search.

### `internal/resource`

Discovers context files, system prompts, skills, prompt templates, themes, and
package resources after trust resolution.

### `internal/pkgmgr`

Installs, removes, updates, lists, filters, enables, and disables resource
packages from local paths, git sources, and archives.

### `internal/share`

Creates private share artifacts and uploads them only after explicit user action.

### `internal/extension`

Runs extension processes and translates lifecycle/tool/UI events over JSONL RPC.
The core never imports extension code.

### `internal/tui`

Owns raw terminal mode, input parsing, renderer diffing, components, focus,
overlays, keybindings, and terminal capability detection.

## First Implementation Slice

The first code slice should prove the architecture, not the whole product:

1. `go.mod`, `cmd/nu`, `internal/app`.
2. CLI parse for print mode.
3. Fake provider in tests.
4. Agent loop with one text response and one tool call.
5. Session append/load.
6. JSON event output.

After that, each Pi feature lands behind a spec ID and tests.
