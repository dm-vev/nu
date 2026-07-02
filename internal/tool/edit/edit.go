package edit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run applies exact replacements to one file under cwd.
func Run(ctx context.Context, cwd string, raw string) (agent.ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Replacements []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"replacements"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if len(args.Replacements) == 0 {
		return agent.ToolResult{}, errors.New("edit: missing replacements")
	}
	path, rel, err := toolkit.ResolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("edit %s: %w", rel, err)
	}

	toolkit.MutationMu.Lock()
	defer toolkit.MutationMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("read %s: %w", rel, err)
	}
	original := string(data)
	edited := original
	var patch strings.Builder
	patch.WriteString("--- ")
	patch.WriteString(rel)
	patch.WriteString("\n+++ ")
	patch.WriteString(rel)
	patch.WriteString("\n@@\n")
	for _, replacement := range args.Replacements {
		if replacement.Old == "" {
			return agent.ToolResult{}, fmt.Errorf("edit %s: empty old text", rel)
		}
		count := strings.Count(original, replacement.Old)
		if count == 0 {
			return agent.ToolResult{}, fmt.Errorf("edit %s: missing old text", rel)
		}
		if count > 1 {
			return agent.ToolResult{}, fmt.Errorf("edit %s: ambiguous old text", rel)
		}
		edited = strings.Replace(edited, replacement.Old, replacement.New, 1)
		patch.WriteString("-")
		patch.WriteString(replacement.Old)
		patch.WriteString("\n+")
		patch.WriteString(replacement.New)
		patch.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return toolkit.JSONResult(map[string]any{"path": rel, "patch": patch.String()})
}
