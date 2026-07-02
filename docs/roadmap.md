# Roadmap

This is implementation order, not product scope. Product scope is in
`spec/functions.md`.

## Phase 0: Project Skeleton

- `go.mod`
- `cmd/nu`
- `internal/app`
- test harness with temp home/cwd
- fake provider
- protocol golden-test fixtures for provider stream, session JSONL, RPC JSONL,
  extension JSONL, and TUI render/input contracts

Exit: `nu --help` works and `go test ./...` passes.

## Phase 1: Headless Agent Spine

Status: implemented through `456582c`.

- CLI parse for print and JSON mode
- message/event types
- provider interface
- agent loop
- session append/load
- one fake tool in tests
- provider stream assembly contract

Exit: a fake provider can stream text, request a tool, receive the result, and
finish with JSON events.

## Phase 2: Built-in Tools

Status: implemented through `3d3fb26`.

- read
- write
- edit
- bash
- grep
- find
- ls
- truncation and full-output persistence
- review fixes for symlink path escapes, multi-edit ordering, long grep lines,
  and Windows bash package compilation

Exit: tools are test-covered without real provider calls.

## Phase 3: Real Providers And Models

Status: implemented through `4ddd508`.

- OpenAI Chat Completions
- OpenAI Responses
- Anthropic Messages
- Google Generative AI
- Bedrock
- OpenAI-compatible custom providers
- auth and model registry

Exit: mocked provider contract tests pass; real smoke tests are opt-in.

## Phase 4: Sessions And Compaction

- resume, continue, fork, clone
- tree storage
- import/export JSONL
- compaction
- branch summaries

Exit: sessions can branch and compact without losing tool-call integrity.

## Phase 5: Interactive TUI

- terminal raw mode
- renderer
- editor
- keybindings
- message view
- footer/status
- selectors
- overlays
- themes

Exit: fake-terminal tests cover input, render width, resize, and basic commands.

## Phase 6: Resources

- context files
- system prompts
- skills
- prompt templates
- packages
- project trust

Exit: trusted project resources load; untrusted resources are blocked.

## Phase 7: Extension Host

- JSONL extension process protocol
- lifecycle hooks
- tool hooks
- command registration
- UI requests
- state persistence
- optional Pi TypeScript compatibility host

Exit: a sample extension registers a tool and blocks a dangerous tool call.

## Spec Correction Rule

Each phase is allowed to split, merge, or rename planned files only by changing
`spec/packages/*` first. Protocol files in `spec/protocols/*` require golden
test updates for any wire-format change.

## Phase 8: RPC, Export, Update

- full RPC commands
- HTML export
- share command
- self/package update
- offline mode checks

Exit: headless clients can drive Nu without TUI.

## Phase 9: Compatibility Sweep

- Pi session import
- Pi settings import
- Pi skills/prompts/themes import
- command parity audit against `third-party/pi`

Exit: every `NUF-*` requirement has passing tests or an explicit follow-up.
