package find

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

// Run finds files under cwd.
func Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Root  string `json:"root"`
		Glob  string `json:"glob"`
		Limit int    `json:"limit"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if args.Limit <= 0 {
		args.Limit = 100
	}
	root, _, err := toolkit.ResolveUnder(cwd, toolkit.DefaultPath(args.Root, "."))
	if err != nil {
		return agent.ToolResult{}, err
	}
	ignore := toolkit.LoadGitignore(root)
	paths := []string{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := toolkit.Relative(cwd, path)
		if err != nil {
			return err
		}
		if toolkit.ShouldSkip(rel, entry.IsDir(), ignore) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !toolkit.GlobMatches(args.Glob, rel) {
			return nil
		}
		paths = append(paths, filepath.ToSlash(rel))
		if len(paths) >= args.Limit {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("find: %w", err)
	}
	sort.Strings(paths)
	return toolkit.JSONListResult("paths", paths, maxBytes)
}
