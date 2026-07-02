package read

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run reads one text or supported image file under cwd.
func Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := toolkit.ResolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("read %s: %w", rel, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if info.IsDir() {
		return agent.ToolResult{}, fmt.Errorf("read %s: is a directory", rel)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("read %s: %w", rel, err)
	}
	if mimeType := toolkit.ImageMime(path); mimeType != "" {
		return toolkit.JSONResult(map[string]any{
			"path":      rel,
			"mime_type": mimeType,
			"data":      base64.StdEncoding.EncodeToString(data),
		})
	}
	if args.Offset < 0 || args.Offset > len(data) {
		return agent.ToolResult{}, fmt.Errorf("read %s: invalid offset", rel)
	}
	data = data[args.Offset:]
	if args.Limit > 0 && args.Limit < len(data) {
		data = data[:args.Limit]
	}
	content, truncated := toolkit.TruncateString(string(data), maxBytes)
	return toolkit.JSONResult(map[string]any{
		"path":      rel,
		"content":   content,
		"truncated": truncated,
	})
}
