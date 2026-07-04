# `internal/tui/components`

## Status

Current: IMPLEMENTED_UNCOMMITTED
Implementation Commit: pending TUI commit; user requested no commits until TUI is complete.

## Purpose

Reusable Pi-style TUI components. Each visible component family is isolated in
its own subpackage so message, header, footer, editor-adjacent, and primitive
rendering behavior can be tested independently.

## Subpackages

- `assistantmessage`: assistant turn rendering.
- `usermessage`: user turn rendering.
- `markdown`: stdlib-only Markdown subset for chat content.
- `thinking`: gray/italic model reasoning blocks.
- `toolblock`: command, patch, and generic tool execution blocks.
- `header`: startup logo/help/onboarding.
- `footer`: cwd/context/model footer.
- `status`: transient status line.
- `fill`: flexible blank space for bottom anchoring.
- `text`: wrapping cached text.
- `box`: padding/background container.
- `border`: horizontal border.
- `spacer`: blank-line spacer.

## Acceptance Criteria

- Components return `[]string`; they never write to the terminal.
- Every component has a package-local test.
- Rendered output is width-bounded or explicitly delegated to `ansi.PadRight`.
