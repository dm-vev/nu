# Nu Spec

Nu is a Vanilla Go coding-agent harness. The goal is not a small Pi clone: Pi is
the minimum functional baseline. Every feature starts here as a spec, then gets
tests, then implementation.

## Source Of Truth

- `spec/product.md`: product scope and non-negotiable behavior.
- `spec/architecture.md`: module boundaries and runtime flows.
- `spec/functions.md`: user-visible functional requirements.
- `spec/protocols/`: persisted and streamed wire-format contracts.
- `spec/packages/`: per-subpackage and per-file implementation contracts.
- `spec/testing.md`: TDD rules and required test shapes.
- `docs/architecture.md`: implementation-oriented architecture notes.
- `docs/development.md`: daily development workflow.

When implementation and spec disagree, the spec wins. Update the spec first
when the intended behavior changes.

## SDD Workflow

1. Write or update a requirement in `spec/functions.md`.
2. Add acceptance criteria and at least one test case name.
3. Re-check the affected `spec/packages/*` files and make their function logic
   concrete enough to implement.
4. Write the failing test.
5. Implement the smallest code that makes the test pass.
6. Run the package test, then the relevant integration test.
7. Update docs only for behavior users or contributors need to know.

## Requirement IDs

Use stable IDs in tests and docs:

- `NUF-*`: functional requirements.
- `NUT-*`: test requirements.
- `NUA-*`: architecture decisions.

Test names should include the requirement ID when practical, for example
`TestNUF080SessionAppendBuildsTree`.

Protocol tests should include the protocol name, for example
`TestProviderStreamAssemblesInterleavedToolCalls`.

## Compatibility Rule

Pi compatibility means behavior compatibility, not TypeScript source
compatibility. Nu must be able to perform the same user jobs and preserve or
import important Pi data formats where useful. A direct TypeScript runtime inside
the Go process is not part of the core architecture.
