# `internal/pkgmgr/manager.go`

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

Install, remove, update, list, filter, enable, and disable resource packages.

## Code Style

Settings mutations are explicit. Package commands return structured results for
TUI/CLI rendering.

## Functions

### `Install(ctx context.Context, opts InstallOptions) (Result, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Install local/git/archive source into selected global or project scope.
- Updates settings with normalized source.

Acceptance:

- installs local/git/archive source into selected global or project scope;
- updates settings with normalized source.

### `Remove(ctx context.Context, opts RemoveOptions) (Result, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Remove matching package from settings.
- Do not delete local path packages.

Acceptance:

- removes matching package from settings;
- does not delete local path packages.

### `Update(ctx context.Context, opts UpdateOptions) (Result, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Resolve selected installed packages from settings and current trust decision.
- For git/archive packages, refresh only the package-owned checkout/cache path.
- Leave settings pointing at the previous resolved package when refresh fails.

Acceptance:

- updates selected installed packages only;
- leaves current package entry untouched on failed refresh.

### `Enable(ctx context.Context, opts EnableOptions) (Result, error)`

Logic:

- Check `ctx` before settings mutation.
- Find package by normalized id in the selected global or project scope.
- Clear the disabled flag without reinstalling or touching package files.
- Return the updated package entry for CLI/TUI rendering.

Acceptance:

- enables an installed package in settings;
- does not reinstall the package.

### `Disable(ctx context.Context, opts DisableOptions) (Result, error)`

Logic:

- Check `ctx` before settings mutation.
- Find package by normalized id in the selected global or project scope.
- Set the disabled flag and leave source/path metadata intact.
- Return the updated package entry for CLI/TUI rendering.

Acceptance:

- disables an installed package in settings;
- keeps package source metadata for later re-enable.

### `List(ctx context.Context, opts ListOptions) ([]ResolvedPackage, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Include global and trusted project packages with project overriding global.

Acceptance:

- includes global and trusted project packages with project overriding global.

Tests:

- `TestNUF150PackageFilterIncludesAndExcludes`
- `TestNUF150PackageEnableDisableUpdatesSettings`
