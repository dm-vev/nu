# Package File Specs

This tree mirrors the future Go source tree. Every planned Go file gets one
`*.go.md` spec file in the matching package folder.

Each file spec must include:

- purpose;
- current status;
- TODO checklist;
- implementation commit and comments;
- code style;
- functions/types owned by the file;
- function logic;
- acceptance criteria;
- tests that prove the contract.

Implementation rule: do not add a Go file or exported function before its spec
exists here.

Readiness rule: do not start implementing a subpackage from the initial tree
blindly. Before the first Go code in a subpackage, re-open every file spec in
that subpackage and confirm:

- file boundaries still match the protocol specs and phase goal;
- each function has concrete step-by-step logic, not acceptance criteria copied
  as implementation;
- tests named in the file prove the risky behavior first;
- any split/merge/rename is applied to `spec/packages/*` before Go files are
  created.

## Status Values

- `TODO`: not implemented.
- `IN_PROGRESS`: implementation branch/worktree has started.
- `IMPLEMENTED`: code exists and tests listed in the file pass.
- `BLOCKED`: cannot proceed without a decision or external dependency.

When a file becomes `IMPLEMENTED`, update that file spec:

- set `Current: IMPLEMENTED`;
- replace `Implementation Commit: -` with the commit hash that introduced or
  completed the implementation;
- replace `Implementation Comments` with short implementation notes, tradeoffs,
  or follow-up risks;
- check off completed TODO items.

If implementation spans multiple commits, record the final commit that made the
listed tests pass and mention earlier commits in comments only if needed.

Common style for all Go files:

- keep package names lowercase and short;
- pass `context.Context` into blocking, streaming, process, network, and storage
  functions;
- return errors, do not panic for recoverable failures;
- wrap errors with operation context;
- avoid global mutable state;
- define interfaces in the consuming package;
- use table tests for parsers, matchers, and protocol conversion;
- keep `cmd/nu` thin.
