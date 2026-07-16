# Agent SDK Examples

Status: **IMPLEMENTED: examples use the balanced package APIs**.

## Purpose

`examples/` contains short, runnable programs that demonstrate the integrated
Agent SDK through its approved balanced API. Examples show the happy path
and are not alternate application layers, test fixtures, or compatibility
wrappers.

Examples compile only against the final packages. They must not retain or hide
imports of the removed root SDK facade, `internal/interfaces`, temporary flat
implementation owners, or any other superseded path. No example-only adapter
may simulate the old API.

## Example Set

| Directory | Balanced API imports | Contract |
|---|---|---|
| `research` | `agent`, `internal/config`, `internal/llm/openai`, `internal/memory/conversation`, `internal/multitenancy`, `telemetry/otel`, `internal/tools/{registry,search}` | Construct and run an OpenAI research agent with scoped conversation context. |
| `providers` | `internal/config`, each applicable `internal/llm/*` provider package | Construct imported providers directly without making model requests. |
| `tools` | `internal/tools/{calculator,registry}` | Register and execute the Calculator through the tool registry. |
| `memory` | `contracts`, `internal/memory/conversation`, `internal/multitenancy` | Store and read messages with organization and conversation context. |
| `mcp` | `internal/mcp/builder` | Build lazy stdio and HTTP configurations without connecting. |
| `task` | `internal/task`, `internal/task/service` | Execute a local task through the task service API. |
| `tracing` | `telemetry/otel` | Create local OpenTelemetry spans without an exporter. |

## Style

- Each example is one small `package main` program.
- Use SDK constructors and options directly; do not add local DTO layers,
  dispatch frameworks, fake streams, or reusable example abstractions.
- Update calls to the final owning package API; preserving an old call shape
  with a forwarding helper is a compatibility wrapper and is forbidden.
- Use normal package-local API names; do not preserve temporary flat-package
  prefixes in example helpers.
- Prefer the direct happy path with ordinary `log.Fatal` error handling.
- Keep credentials in environment variables and never print them.
- Only `research` performs provider I/O. Other examples must run locally without
  credentials, subprocesses, or network access.
- Examples remain inside this module because the integrated SDK is under Go's
  `internal/` boundary.

## Verification

```bash
go test ./examples/...
go run ./examples/providers
go run ./examples/tools
go run ./examples/memory
go run ./examples/mcp
go run ./examples/task
go run ./examples/tracing
```

`research` is compile-tested by the first command and is run manually only with
valid provider credentials.
