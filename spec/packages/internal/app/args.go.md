# `internal/app/args.go`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: -
Implementation Comments: Parser records provider/model/auth/model-file selectors and preserves extension flags in the app owner.

## TODO

- [x] Add or confirm the failing tests listed in this file.
- [x] Implement the file according to the function logic below.
- [x] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Parse Nu/Pi-compatible CLI flags into a typed request.

## Code Style

Use stdlib parsing or a tiny hand parser. Preserve unknown long flags for
extensions. Do not call `os.Exit`.

## Functions

### `Parse(args []string) (Request, Diagnostics)`

Logic:

- Iterate argv left-to-right and classify tokens as known flags, unknown extension flags, file references, or prompt text.
- Parse known flags into typed `Request` fields and attach source spans for diagnostics.
- Store `--provider`, `--model`, `--api-key`, and `--models` values in the
  request instead of discarding them.
- Keep `@file` references separate from prompt text so resource loading can resolve them later.
- Record invalid values as diagnostics and continue parsing when the remaining argv is still meaningful.

Acceptance:

- supports all flags listed by `NUF-001`;
- keeps `@file` arguments separate from prompt messages;
- records invalid values as diagnostics;
- preserves unknown `--flag` values for extension flags.

### `parseThinkingLevel(value string) (model.ThinkingLevel, bool)`

Logic:

- Trim and lowercase the raw flag value.
- Accept only `off`, `minimal`, `low`, `medium`, `high`, and `xhigh`.
- Return `false` for unknown values without selecting a default.
- Keep provider-specific thinking mapping in `internal/model/thinking.go`.

Acceptance:

- accepts only `off`, `minimal`, `low`, `medium`, `high`, `xhigh`.

Tests:

- `TestNUF001ParseKnownFlags`
- `TestNUF001UnknownFlagsArePreservedForExtensions`
- `TestNUF001InvalidThinkingLevelReportsDiagnostic`
