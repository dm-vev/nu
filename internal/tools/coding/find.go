package coding

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Run finds files under cwd.
func RunFind(ctx context.Context, cwd string, raw string, maxBytes int) (Result, error) {
	var args struct {
		Root  string `json:"root"`
		Glob  string `json:"glob"`
		Limit int    `json:"limit"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return Result{}, err
	}
	if args.Limit <= 0 {
		args.Limit = 100
	}
	root, _, err := resolveUnder(cwd, defaultPath(args.Root, "."))
	if err != nil {
		return Result{}, err
	}
	ignore := loadGitignore(root)
	paths := []string{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := relative(cwd, path)
		if err != nil {
			return err
		}
		if shouldSkip(rel, entry.IsDir(), ignore) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !globMatches(args.Glob, rel) {
			return nil
		}
		paths = append(paths, filepath.ToSlash(rel))
		if len(paths) >= args.Limit {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("find: %w", err)
	}
	sort.Strings(paths)
	return jsonListResult("paths", paths, maxBytes)
}
