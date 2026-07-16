# `internal/app/file_structure_test.go`

## Status

Current: IN_PROGRESS
Implementation Commit: -

## Purpose

Enforce NUT-009's repository-wide 300-line limit and the balanced hierarchy's
production package allowlist without creating a test-only root package.

## Tests

- `TestNUT009ProductionGoFilesAreAtMost300Lines` reports oversized production
  files while excluding generated and test files.
- `TestNUF212HierarchyHasNoOldPackagesOrFacade` rejects superseded paths,
  wrappers, and any `agent` dependency on concrete task/transport packages.
- `TestNUA009InternalPackagesMatchBalancedHierarchy` requires every exact
  root/subpackage/standalone owner from NUA-011, including the seven
  `internal/llm/*` provider packages and four `internal/data/*` packages, and
  rejects all others.
- `TestNUF204ToolRootDoesNotReExportChildPackages` keeps root tool orchestration
  independent from `coding`, `search`, `image`, and `graphrag` implementations.
- The structure check also rejects nested component packages, provider-prefixed
  subpackage filenames, and one-helper packages without a cohesive boundary.
