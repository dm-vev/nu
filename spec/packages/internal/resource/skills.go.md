# `internal/resource/skills.go`

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

Discover, parse, validate, and expose skills.

## Code Style

Read only frontmatter and summary during discovery. Full content is loaded on
demand.

## Functions

### `DiscoverSkills(ctx context.Context, opts SkillOptions) ([]Skill, Diagnostics, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Search global, project, package, settings, and CLI locations.
- Validate required name/description.
- Warn on collisions and keeps first.

Acceptance:

- searches global, project, package, settings, and CLI locations;
- validates required name/description;
- warns on collisions and keeps first.

### `LoadSkill(ctx context.Context, skill Skill) (SkillContent, error)`

Logic:

- Check `ctx` before blocking work and pass it to every blocking dependency.
- Enumerate configured sources in precedence order.
- Parse each source independently and aggregate diagnostics.
- Read full `SKILL.md`.
- Resolve relative asset/script paths from skill directory.

Acceptance:

- reads full `SKILL.md`;
- resolves relative asset/script paths from skill directory.

Tests:

- `TestNUF130SkillDiscovery`
- `TestNUF130SkillCommandExpandsContent`
