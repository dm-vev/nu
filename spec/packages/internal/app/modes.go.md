# `internal/app/modes.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -
Implementation Comments: Help/version/print dispatch exists; JSON/RPC/package/share/update modes are pending.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Dispatch resolved CLI requests to interactive, print, JSON, RPC, package,
config, export, share, and update modes.

## Code Style

Use a small switch over command kind. Mode handlers return `error`; conversion
to exit code stays in `app.go`.

## Functions

### `runMode(ctx context.Context, rt *Runtime, req cli.Request) error`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Initialize local state, then enter the smallest required loop.
- Stop on context cancellation, terminal command, or unrecoverable error and clean up owned resources.
- Dispatches every `NUF-002` mode.
- Refuse incompatible flag combinations with a clear error.
- Preserve stdout JSONL purity for JSON/RPC modes.

Acceptance:

- dispatches every `NUF-002` mode;
- refuses incompatible flag combinations with a clear error;
- preserves stdout JSONL purity for JSON/RPC modes.

Tests:

- `TestNUF002DispatchPrintMode`
- `TestNUF002DispatchRPCMode`
- `TestNUF002DispatchPackageCommand`
- `TestNUF002DispatchShareCommand`
