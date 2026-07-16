package coding

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
)

// Run reads one text or supported image file under cwd.
func RunRead(ctx context.Context, cwd string, raw string, maxBytes int) (Result, error) {
	var args struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return Result{}, err
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("read %s: %w", rel, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if info.IsDir() {
		return Result{}, fmt.Errorf("read %s: is a directory", rel)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Result{}, fmt.Errorf("read %s: %w", rel, err)
	}
	if mimeType := imageMIME(path); mimeType != "" {
		return jsonResult(map[string]any{
			"path":      rel,
			"mime_type": mimeType,
			"data":      base64.StdEncoding.EncodeToString(data),
		})
	}
	if args.Offset < 0 || args.Offset > len(data) {
		return Result{}, fmt.Errorf("read %s: invalid offset", rel)
	}
	data = data[args.Offset:]
	if args.Limit > 0 && args.Limit < len(data) {
		data = data[:args.Limit]
	}
	content, truncated := truncateString(string(data), maxBytes)
	return jsonResult(map[string]any{
		"path":      rel,
		"content":   content,
		"truncated": truncated,
	})
}
