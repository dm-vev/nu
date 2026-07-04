package tool

import (
	"context"

	"nu/internal/agent"
	"nu/internal/provider"
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

// Definitions returns provider-facing schemas for built-in tools.
func Definitions() []provider.ToolDefinition {
	return []provider.ToolDefinition{
		{
			Name:        "bash",
			Description: "Run a shell command in the current project directory and return stdout, stderr, and exit code.",
			Parameters: objectSchema(map[string]any{
				"command":    map[string]any{"type": "string", "description": "Shell command to execute."},
				"timeout_ms": map[string]any{"type": "integer", "description": "Optional timeout in milliseconds."},
			}, []string{"command"}),
		},
		{
			Name:        "read",
			Description: "Read a text file under the current project directory.",
			Parameters: objectSchema(map[string]any{
				"path":   map[string]any{"type": "string"},
				"offset": map[string]any{"type": "integer"},
				"limit":  map[string]any{"type": "integer"},
			}, []string{"path"}),
		},
		{
			Name:        "write",
			Description: "Create or overwrite a file under the current project directory.",
			Parameters: objectSchema(map[string]any{
				"path":    map[string]any{"type": "string"},
				"content": map[string]any{"type": "string"},
			}, []string{"path", "content"}),
		},
		{
			Name:        "edit",
			Description: "Apply exact text replacements to a file and return a patch.",
			Parameters: objectSchema(map[string]any{
				"path": map[string]any{"type": "string"},
				"replacements": map[string]any{
					"type": "array",
					"items": objectSchema(map[string]any{
						"old": map[string]any{"type": "string"},
						"new": map[string]any{"type": "string"},
					}, []string{"old", "new"}),
				},
			}, []string{"path", "replacements"}),
		},
		{
			Name:        "grep",
			Description: "Search files under the current project directory for a text pattern.",
			Parameters: objectSchema(map[string]any{
				"pattern": map[string]any{"type": "string"},
				"path":    map[string]any{"type": "string"},
			}, []string{"pattern"}),
		},
		{
			Name:        "find",
			Description: "Find files under the current project directory by name pattern.",
			Parameters: objectSchema(map[string]any{
				"name": map[string]any{"type": "string"},
				"path": map[string]any{"type": "string"},
			}, nil),
		},
		{
			Name:        "ls",
			Description: "List a directory under the current project directory.",
			Parameters: objectSchema(map[string]any{
				"path": map[string]any{"type": "string"},
			}, nil),
		},
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}
