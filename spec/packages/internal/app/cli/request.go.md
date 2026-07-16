# `internal/app/cli/request.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Request shape carries provider/model/auth/model-file selectors in the app owner; detailed session/resource fields are pending.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Define parsed CLI request types.

## Code Style

Types only. Avoid helper methods unless they remove repeated branching in
consumers.

## Types

### `type Request struct`

Logic:

- Store command kind, execution mode, model selector, auth selector, session reference, tool filters, resource flags, prompt messages, file arguments, raw extension flags, and diagnostics source spans.
- Store `provider`, `model`, `api_key`, and `models` flag values for Phase 3
  provider/runtime selection.
- Keep process IO, filesystem, environment, and runtime dependencies out of this type.
- Keep validation in `args.go` and mode dispatch in `modes.go` so this file remains a pure request contract.

Acceptance:

- represents command kind, mode, model/auth/session/tool/resource flags, prompt
  messages, file args, and extension flags.

### `type CommandKind string`

Logic:

- Define the closed set of top-level commands used by `app/modes.go`: chat,
  package, config, update, export, share, help, version, and list-models.
- Keep string values stable because JSON/RPC mode and scripted output may expose them.
- Add new command kinds only after updating parser, help, mode dispatch, and tests.

Acceptance:

- covers chat, package, config, update, export, share, help, version, and
  list-models.

## Functions

No parser or renderer functions belong in this file. Parsing stays in
`args.go`; help rendering stays in `help.go`.

Tests:

- covered by `args.go` parser tests.
