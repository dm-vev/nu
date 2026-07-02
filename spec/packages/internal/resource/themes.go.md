# `internal/resource/themes.go`

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

Load built-in and custom TUI themes.

## Code Style

Theme parsing validates colors and required fields. Broken themes become
diagnostics, not startup crashes.

## Functions

### `LoadThemes(ctx context.Context, opts ThemeOptions) ([]Theme, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Include dark and light built-ins.
- Load global, project, package, settings, and CLI themes.
- De-duplicate by name.

Acceptance:

- includes dark and light built-ins;
- loads global, project, package, settings, and CLI themes;
- de-duplicates by name.

Tests:

- `TestNUF103ThemeLoadsFromResource`
- `TestNUF103BrokenThemeReportsDiagnostic`
