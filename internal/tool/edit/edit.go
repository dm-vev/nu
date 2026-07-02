package edit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
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
	type span struct {
		start int
		end   int
		new   string
	}
	spans := make([]span, 0, len(args.Replacements))
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
		start := strings.Index(original, replacement.Old)
		if start < 0 {
			return agent.ToolResult{}, fmt.Errorf("edit %s: missing old text", rel)
		}
		if strings.Count(original, replacement.Old) > 1 {
			return agent.ToolResult{}, fmt.Errorf("edit %s: ambiguous old text", rel)
		}
		spans = append(spans, span{
			start: start,
			end:   start + len(replacement.Old),
			new:   replacement.New,
		})
		patch.WriteString("-")
		patch.WriteString(replacement.Old)
		patch.WriteString("\n+")
		patch.WriteString(replacement.New)
		patch.WriteString("\n")
	}
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].start < spans[j].start
	})
	for i := 1; i < len(spans); i++ {
		if spans[i-1].end > spans[i].start {
			return agent.ToolResult{}, fmt.Errorf("edit %s: overlapping replacements", rel)
		}
	}
	edited := original
	for i := len(spans) - 1; i >= 0; i-- {
		// Apply against original byte spans so replacement order cannot affect later matches.
		edited = edited[:spans[i].start] + spans[i].new + edited[spans[i].end:]
	}
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return toolkit.JSONResult(map[string]any{"path": rel, "patch": patch.String()})
}
