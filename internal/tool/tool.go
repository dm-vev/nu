package tool

import (
	"context"

	"nu/internal/agent"
	"nu/internal/tool/bash"
	"nu/internal/tool/edit"
	"nu/internal/tool/find"
	"nu/internal/tool/grep"
	"nu/internal/tool/ls"
	"nu/internal/tool/read"
	"nu/internal/tool/write"
)

const defaultMaxOutputBytes = 16 * 1024

// Builtins returns the Phase 2 built-in tool set.
func Builtins(cwd string) map[string]agent.ToolFunc {
	return map[string]agent.ToolFunc{
		"read": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return read.Run(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"write": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return write.Run(ctx, cwd, call.Arguments)
		},
		"edit": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return edit.Run(ctx, cwd, call.Arguments)
		},
		"bash": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return bash.Run(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"grep": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return grep.Run(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"find": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return find.Run(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"ls": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return ls.Run(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
	}
}
