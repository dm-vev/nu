package coding

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Run applies exact replacements to one file under cwd.
func RunEdit(ctx context.Context, cwd string, raw string) (Result, error) {
	var args struct {
		Path         string `json:"path"`
		Replacements []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"replacements"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return Result{}, err
	}
	if len(args.Replacements) == 0 {
		return Result{}, errors.New("edit: missing replacements")
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("edit %s: %w", rel, err)
	}

	mutationMu.Lock()
	defer mutationMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read %s: %w", rel, err)
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
			return Result{}, fmt.Errorf("edit %s: empty old text", rel)
		}
		start := strings.Index(original, replacement.Old)
		if start < 0 {
			return Result{}, fmt.Errorf("edit %s: missing old text", rel)
		}
		if strings.Count(original, replacement.Old) > 1 {
			return Result{}, fmt.Errorf("edit %s: ambiguous old text", rel)
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
			return Result{}, fmt.Errorf("edit %s: overlapping replacements", rel)
		}
	}
	edited := original
	for i := len(spans) - 1; i >= 0; i-- {
		// Apply against original byte spans so replacement order cannot affect later matches.
		edited = edited[:spans[i].start] + spans[i].new + edited[spans[i].end:]
	}
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		return Result{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return jsonResult(map[string]any{"path": rel, "patch": patch.String()})
}
