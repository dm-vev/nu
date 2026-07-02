# `internal/agent/tools.go`

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

Validate and execute tool calls for the agent loop.

## Code Style

Tool scheduling stays here; tool implementation stays in `internal/tool`.

## Functions

### `executeToolCalls(ctx context.Context, calls []message.ToolCall, tools tool.Registry, policy ToolPolicy) ([]message.ToolResult, error)`

Logic:

- Allocate a result slice with the same length as `calls`.
- For each call, resolve tool by exact name and validate arguments against the
  tool schema before any execution begins.
- If validation fails, produce a tool-result error at the same index instead of
  skipping the call.
- Choose sequential execution when global policy is sequential or any resolved
  tool requires sequential mode.
- Sequential mode executes in call order and stops only when context is
  cancelled; tool errors are results, not loop errors.
- Parallel mode starts one goroutine per valid call, writes each result into its
  original index, waits for all goroutines, and respects context cancellation.
- Emit tool start/update/end events through the agent event sink.

Acceptance:

- validates tool names and JSON args before execution;
- runs parallel by default;
- runs sequential when policy or tool requires it;
- preserves provider tool call order in returned results.

### `executeOneTool(ctx context.Context, call message.ToolCall, t tool.Tool) message.ToolResult`

Logic:

- Build typed execution context with cwd, session id, tool call id, update
  callback, and extension metadata.
- Call tool `Execute` with `ctx` and decoded arguments.
- Convert returned tool result to `message.ToolResult` while preserving tool
  call id/name.
- Convert tool validation/execution errors into `is_error=true` tool results.
- Recover panics from tool implementation and record them as internal tool
  errors with stack omitted from normal user output.

Acceptance:

- converts panics to tool errors only if a tool violates contract;
- never drops tool call id.

Tests:

- `TestNUF051ParallelToolCallsPreserveResultOrder`
- `TestNUF051SequentialToolRunsInOrder`
- `TestNUF051InvalidArgsReturnToolError`
