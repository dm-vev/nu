# Repository Guidelines

## Project Structure & Module Organization

Nu is currently spec- and documentation-led. Treat `spec/` as the source of
truth for required behavior and `docs/` as contributor-facing implementation
guidance.

- `spec/product.md`: product scope and non-negotiable behavior.
- `spec/architecture.md`: module boundaries, runtime flows, and dependency notes.
- `spec/functions.md`: user-visible functional requirements.
- `spec/testing.md`: required test shapes and test isolation rules.
- `docs/architecture.md`: implementation-oriented architecture notes.
- `docs/development.md`: daily workflow and command reference.

When implementation and spec disagree, update or follow the spec first.

## Build, Test, and Development Commands

Use the documented Go workflow once source packages exist:

```bash
go test ./...
go test -race ./internal/agent ./internal/session ./internal/tool
go vet ./...
gofmt -w <files>
```

Run narrow package tests while developing, then `go test ./...` before handoff.
Docs-only edits do not require Go tests.

## Coding Style & Naming Conventions

Prefer plain Go and the standard library. Add dependencies only when stdlib or
native platform support is not enough, and document each new dependency in
`spec/architecture.md`. Format Go files with `gofmt`; avoid speculative
abstractions and keep package APIs small.

Inside Go functions, add short intent comments before non-trivial control flow,
protocol/state transitions, locks, filesystem/process/network side effects, and
deliberate simplifications. Do not comment obvious assignments; comments must
explain why this step exists or what invariant it protects.

Use stable requirement IDs in specs, docs, and tests:

- `NUF-*` for functional requirements.
- `NUT-*` for testing requirements.
- `NUA-*` for architecture decisions.

Test names should include the requirement ID when practical, for example
`TestNUF080SessionAppendBuildsTree`.

## Testing Guidelines

Use Go's stdlib `testing` package. Default tests must not call real providers,
read real `~/.nu` or `~/.pi`, or depend on provider credentials. Use temp
directories, explicit environment maps, `httptest.Server`, fake terminals, and
package-local `testdata/` fixtures.

For non-trivial changes, follow TDD: update `spec/functions.md`, re-check the
affected `spec/packages/*` files, write a failing test, implement the smallest
passing change, then run the relevant narrow test and final full test command.

## Commit & Pull Request Guidelines

Use short imperative commit subjects such as `add session jsonl tests` or
`document provider contract`.

Pull requests should include the changed requirement IDs, a concise behavior
summary, tests run, and screenshots or terminal output when CLI/TUI behavior
changes. Link related issues when available.
