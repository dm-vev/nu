package coding

import (
	"context"
	"fmt"
	"os"
	"sort"
)

// Run lists one directory under cwd.
func RunLS(ctx context.Context, cwd string, raw string, maxBytes int) (Result, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return Result{}, err
	}
	path, rel, err := resolveUnder(cwd, defaultPath(args.Path, "."))
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("ls %s: %w", rel, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("ls %s: not a directory", rel)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return Result{}, fmt.Errorf("ls %s: %w", rel, err)
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
	return jsonListResult("entries", names, maxBytes)
}
