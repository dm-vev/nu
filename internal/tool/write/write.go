package write

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run creates or overwrites one file under cwd.
func Run(ctx context.Context, cwd string, raw string) (agent.ToolResult, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := toolkit.ResolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}

	toolkit.MutationMu.Lock()
	defer toolkit.MutationMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return agent.ToolResult{}, fmt.Errorf("create parent for %s: %w", rel, err)
	}
	if err := os.WriteFile(path, []byte(args.Content), 0o644); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return toolkit.JSONResult(map[string]any{"path": rel, "bytes": len(args.Content)})
}
