# `internal/app/sdkagent.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Imported SDK Agent construction and Nu agentui adaptation were split from runtime state.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Build the imported SDK agent and wrap it in Nu's event/lifecycle controller.

## Code Style

Use injected dependencies first; keep SDK logging silent on protocol stdout.

## Owned Logic

- `newAgent` defaults tools and memory, accepts an injected runner, and installs model-rebuild support when `BuildLLM` exists.
- `newSDKAgent` configures memory, tools, 16 iterations, intermediate stream content, no plan approval, name `nu`, and a discard logger.
- `discardSDKLogger` satisfies the SDK logger contract without output.

## Acceptance

- No controller is returned without an injected runner or LLM.
- Model rebuilds reuse memory and tools.
- SDK diagnostics cannot corrupt print/JSON/RPC stdout.

## Tests

- `TestAppRunPrintModeUsesInjectedRuntime`
- `TestPrintModeBuildsProviderFromCLI`
- `TestNUF170JSONModeStdoutIsOnlyJSONL`
