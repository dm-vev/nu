# `internal/tools/coding/builtins.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: The seven coding tools are composed in `internal/tools/coding`.

## TODO

- [x] Match the implemented file boundary.
- [x] Confirm package tests cover the owned behavior.
- [x] Run the targeted package tests.
- [ ] Record the implementation commit after commit.

## Purpose

Expose Nu's seven coding tools through the imported SDK `contracts.Tool` contract.

## Code Style

Keep construction side-effect free and delegate execution to package-local `Run*` functions.

## Owned Logic

- `codingTool` stores SDK name, description, parameters, and execution closure.
- `Run`/`Execute` return the JSON result content from the owned tool implementation.
- `Builtins` returns exactly bash, read, write, edit, grep, find, and ls rooted at cwd.

## Acceptance

- Tool construction performs no IO.
- Schemas match each implementation's JSON arguments.
- All seven tools satisfy `contracts.Tool`.

## Tests

- `TestBuiltinsExposesEveryPhaseTwoTool`
- `TestDefinitionsExposeBashSchema`
- `TestBuiltinsReadToolRuns`
