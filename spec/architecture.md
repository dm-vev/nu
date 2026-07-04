# Architecture Spec

## Principles

- One Go module, one binary first.
- `cmd/nu` only parses process startup and calls `internal/app`.
- All product code lives under `internal/`.
- Interfaces are defined at consumers, not providers.
- Context cancellation flows through provider calls, tools, subprocesses, RPC,
  TUI waits, and extension calls.
- Formats are JSON or JSONL unless a terminal protocol requires escape codes.
- Dependencies need a spec note before they are added.
- Protocol specs are stronger than file specs. If a package spec conflicts with
  `spec/protocols/*`, the protocol spec wins.
- Planned file boundaries may change only by updating `spec/packages/*` first.

## Package Map

```text
cmd/nu                         process entry point
internal/app                   composition root and lifecycle
internal/cli                   args, help, command dispatch
internal/config                settings, paths, trust, keybindings
internal/auth                  auth.json, env resolution, OAuth tokens
internal/model                 model registry, matching, thinking levels
internal/provider              provider interface and shared request model
internal/provider/openai       OpenAI Chat Completions and Responses
internal/provider/anthropic    Anthropic Messages
internal/provider/google       Gemini-compatible API
internal/provider/bedrock      AWS Bedrock
internal/provider/compat       OpenAI-compatible custom providers
internal/agent                 turn loop, tool loop, retry, queues
internal/session               JSONL tree storage, resume, fork, clone
internal/compact               compaction and branch summarization
internal/tool                  tool registry, schemas, execution contracts
internal/tool/builtin          read/write/edit/bash/grep/find/ls
internal/resource              context files, skills, prompts, themes
internal/pkgmgr                git/local/npm-like package source handling
internal/extension             process extension host and event hooks
internal/tui                   app wiring plus terminal/editor/render subpackages
internal/slash                 slash command parsing and handlers
internal/rpc                   JSONL RPC server/client protocol
internal/export                JSONL and HTML export
internal/share                 explicit private share artifact upload
internal/update                self/package update commands
internal/platform              OS-specific terminal, clipboard, paths
internal/testkit               fake providers, fake terminal, fixtures
```

## Runtime Flow

### Startup

1. `cmd/nu` starts `internal/app`.
2. `cli` parses args into a command request.
3. `config` resolves global/project settings and trust.
4. `auth` loads credentials and environment overrides.
5. `resource` discovers context files, skills, prompts, themes, packages, and
   extensions allowed by trust and flags.
6. `model` builds the available model registry.
7. `app` starts interactive, print, JSON, RPC, package, update, export, or
   config mode.

### Agent Turn

1. User input becomes a `message.User`.
2. `agent` builds provider context from system prompt, summaries, session path,
   queued messages, active tools, and settings.
3. `provider` streams assistant events.
4. `agent` converts stream deltas into session events.
5. Tool calls are validated, passed through extension hooks, executed, and
   appended as tool result messages.
6. The loop continues until the provider stops without tool calls, aborts, or
   fails past retry policy.
7. Session state is flushed after every durable event.

### Tool Execution

Tools expose:

- name;
- description;
- JSON schema;
- execution mode: parallel or sequential;
- execute function with context cancellation;
- optional renderer metadata for TUI and export.

Built-in write/edit tools serialize file mutations per path. Bash keeps bounded
memory and writes full truncated output to a temp file.

### Session Tree

Session entries are append-only JSONL records with `id`, `parent_id`, timestamp,
type, and payload. The active branch is the path from root to current leaf.
Branching never rewrites existing entries.

### Extensions

Nu extensions run out-of-process over JSONL RPC. They can register tools,
commands, keybindings, resource providers, renderers, and event hooks. This
keeps the Go core independent from any language runtime. A Node compatibility
host may later adapt Pi TypeScript extensions to this protocol.

### TUI

The TUI owns terminal raw mode, rendering, input decoding, focus-capable editor
state, and reusable components. Agent logic never writes directly to the
terminal; it emits events that the TUI renders.

The implemented TUI layer follows Pi's package split in Go:

- `internal/tui` wires Nu app state, agent events, prompt submission, and loop
  lifecycle;
- `internal/tui/engine` renders component trees and writes synchronized
  full/differential terminal updates;
- `internal/tui/terminal` owns raw mode, terminal size, resize watching, and
  terminal writes;
- `internal/tui/input` decodes UTF-8, escape sequences, and bracketed paste;
- `internal/tui/editor` owns multiline buffer mutation and bordered rendering;
- `internal/tui/message` stores ordered chat parts for text, thinking, and
  tool executions;
- `internal/tui/components/*` owns one component family per subpackage,
  including Markdown, thinking, and tool execution blocks.

Selectors and extension UI components land as additional component subpackages
instead of being folded into the top-level app wiring.

### RPC

The RPC server owns strict JSONL framing over injected stdin/stdout. It
recognizes Pi command names, keeps stdout protocol-only, forwards agent events,
and stores lightweight runtime state. Commands whose durable backends are still
landing return stable structured responses over in-memory state instead of
printing human text.

## Architecture Decisions

### NUA-001: Process Extensions

Use process-based extensions instead of in-process plugins. Go plugins are not
portable enough, and embedding a JS runtime makes Node a de facto dependency.
The extension protocol must be documented and tested as JSONL.

### NUA-002: Native Session Format With Import

Nu uses a native JSONL schema versioned independently from Pi. It must include a
Pi importer for important existing sessions instead of locking all future
storage decisions to Pi internals.

### NUA-003: Standard Library First

Use stdlib for JSON, HTTP, subprocesses, filesystem, templates, archives, and
testing. Add a dependency only when platform APIs or protocol correctness make
stdlib-only code worse.

### NUA-004: Protocol-First Vertical Slices

The first implementation slices must prove protocol boundaries before filling
every planned package. Provider stream, session JSONL, RPC JSONL, extension
JSONL, and TUI rendering/input contracts get golden tests before feature work
depends on them.

### NUA-005: File Specs Are Contracts, Not Frozen Topology

`spec/packages/*` defines the intended file layout. If implementation proves a
file should split, merge, or move, update that file spec and any affected package
spec before changing Go code. The status and implementation commit fields make
that change auditable.
