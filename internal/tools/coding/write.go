package coding

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Run creates or overwrites one file under cwd.
func RunWrite(ctx context.Context, cwd string, raw string) (Result, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return Result{}, err
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("write %s: %w", rel, err)
	}

	mutationMu.Lock()
	defer mutationMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Result{}, fmt.Errorf("create parent for %s: %w", rel, err)
	}
	if err := os.WriteFile(path, []byte(args.Content), 0o644); err != nil {
		return Result{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return jsonResult(map[string]any{"path": rel, "bytes": len(args.Content)})
}
