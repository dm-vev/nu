# `internal/extension/hooks.go`

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

Dispatch lifecycle, session, model, and tool hooks to extensions.

## Code Style

Hooks are ordered by load order. A blocking hook result short-circuits later
hooks only where spec says it should.

## Functions

### `DispatchToolCall(ctx context.Context, hooks []Hook, ev ToolCallEvent) (ToolDecision, error)`

Logic:

- Build hook request payload with tool name, call id, args, cwd, mode, and
  session metadata.
- Send hooks one by one in extension load order.
- Apply hook timeout per extension.
- On `continue`, keep current payload.
- On `modify`, validate modified args against tool schema before continuing.
- On `block`, return decision immediately with extension id and reason.
- On hook process failure, follow configured policy: fail-closed for security
  hooks, fail-open for informational hooks.

Acceptance:

- allows hooks to approve, modify, or block tool calls;
- includes blocking reason in tool error.

### `DispatchLifecycle(ctx context.Context, hooks []Hook, ev LifecycleEvent) error`

Logic:

- Build lifecycle event once.
- Send to all subscribed extensions in load order.
- Collect diagnostics for non-blocking hook failures.
- Stop early only on context cancellation or a lifecycle event that explicitly
  allows blocking.

Acceptance:

- runs all hooks unless context is canceled.

Tests:

- `TestNUF160ExtensionCanBlockToolCall`
