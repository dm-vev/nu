package ls

import (
	"context"
	"fmt"
	"os"
	"sort"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run lists one directory under cwd.
func Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := toolkit.ResolveUnder(cwd, toolkit.DefaultPath(args.Path, "."))
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("ls %s: %w", rel, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if !info.IsDir() {
		return agent.ToolResult{}, fmt.Errorf("ls %s: not a directory", rel)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("ls %s: %w", rel, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return toolkit.JSONListResult("entries", names, maxBytes)
}
