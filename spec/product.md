# Product Spec

## Goal

Nu is a local-first coding agent written in Vanilla Go. It provides interactive
TUI, print, JSON event stream, and RPC modes; supports multiple model providers;
executes coding tools; persists sessions as branchable trees; and loads user and
project resources such as context files, skills, prompt templates, themes,
packages, and extensions.

## Baseline

Pi is the minimum product baseline:

- interactive coding-agent CLI;
- provider/model/auth management;
- tool calling with read/write/edit/bash/grep/find/ls;
- session persistence, resume, fork, clone, tree navigation;
- compaction and branch summarization;
- slash commands;
- TUI editor, selectors, overlays, keybindings, themes;
- JSON and RPC headless modes;
- resource loading for context files, skills, prompt templates, packages,
  extensions, themes;
- project trust and config isolation;
- export/share/update flows.

Nu may change internal formats only when it keeps migration/import paths.

## Non Goals

- No Node/TypeScript runtime as a hard dependency of the core binary.
- No framework-first rewrite. Standard library comes first.
- No speculative cloud service. Nu runs locally and stores local state by
  default.
- No silent network telemetry. Any analytics must be explicit opt-in.

## Distribution

The primary artifact is a single `nu` binary. Optional helpers may exist for
compatibility shims, but the core agent must start, run tools, and talk to model
providers without Node.

## Persistence Paths

Default paths:

- global config: `~/.nu/agent/`
- project config: `.nu/`
- sessions: `~/.nu/agent/sessions/`
- auth: `~/.nu/agent/auth.json`
- trust: `~/.nu/agent/trust.json`

Pi import paths:

- `~/.pi/agent/`
- `.pi/`

Import must be explicit unless a command says it is a compatibility mode.

## Security Baseline

Nu runs with the user's permissions. It must make that explicit. Project-local
settings, packages, extensions, and executable resources require trust. Auth
files must be written with user-only permissions where the platform supports it.

## Done Definition

A feature is not done until:

- the requirement is written in `spec/functions.md`;
- tests cover the main behavior and one failure path;
- user-facing docs are updated when commands or files are affected;
- `go test ./...` passes;
- `go vet ./...` passes before release.
