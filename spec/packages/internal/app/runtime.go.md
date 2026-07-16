# `internal/app/runtime.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Runtime state and IO normalization were split from provider, model, settings, output, and SDK-agent wiring.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Define the injectable options and shared runtime state for one Nu invocation.

## Code Style

Keep this file dependency-only: no provider construction, filesystem access, or mode behavior.

## Owned Logic

- `Options` carries process IO, paths, model metadata, injected SDK contracts, tools, memory, and session identity.
- `Runtime` holds normalized options shared by mode handlers.
- `normalizeOptions` replaces nil stdin/stdout/stderr with safe empty or discard values.

## Acceptance

- Tests can construct a runtime without process globals.
- Nil IO never causes mode dispatch to panic.

## Tests

- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestNUF002DispatchRPCMode`
- `TestNUF002DispatchInteractiveMode`
