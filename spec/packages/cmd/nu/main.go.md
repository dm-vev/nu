# `cmd/nu/main.go`

## Status

Current: TODO
Implementation Commit: -
Implementation Comments: Not implemented yet.

## TODO

- [ ] Add or confirm the failing tests listed in this file.
- [ ] Implement the file according to the function logic below.
- [ ] Run the targeted package tests.
- [ ] After implementation commit, replace `Implementation Commit` with the commit hash and summarize important comments.

## Purpose

Process entry point for the `nu` binary.

## Code Style

No business logic. No package globals except build-time version variables. Main
creates context, calls `app.Run`, prints fatal errors to stderr, exits with the
returned code.

## Functions

### `main()`

Logic:

- Create a cancellable process context from OS signals.
- Read `os.Args[1:]`, `os.Environ`, cwd, home, stdin, stdout, stderr, and
  build-time version metadata once at process startup.
- Pass those values into `app.Run` through `app.Options`; no business logic or
  config resolution stays in `main`.
- Convert interrupt signals into context cancellation and let the app layer
  decide cleanup and exit code.
- Call `os.Exit` exactly once with the returned app exit code.

Acceptance:

- passes `os.Args[1:]`, stdin, stdout, stderr, and environment to app layer;
- handles interrupt cancellation;
- exits with app exit code.

Tests:

- covered through `internal/app` integration tests, not direct `main` tests.
