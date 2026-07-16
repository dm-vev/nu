package coding

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Result is the Nu coding-tool result adapted to the SDK Tool interface.
type Result struct {
	Content string
}

func decodeArgs(raw string, out any) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return fmt.Errorf("decode tool args: %w", err)
	}
	return nil
}

func truncateString(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value, false
	}
	return value[:maxBytes], true
}

func jsonResult(value map[string]any) (Result, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return Result{}, fmt.Errorf("marshal tool result: %w", err)
	}
	return Result{Content: string(data)}, nil
}

func jsonListResult(key string, values []string, maxBytes int) (Result, error) {
	if values == nil {
		values = []string{}
	}
	truncated := false
	for {
		result := map[string]any{key: values, "truncated": truncated}
		data, err := json.Marshal(result)
		if err != nil {
			return Result{}, fmt.Errorf("marshal tool result: %w", err)
		}
		if maxBytes <= 0 || len(data) <= maxBytes || len(values) == 0 {
			return Result{Content: string(data)}, nil
		}
		truncated = true
		values = values[:len(values)-1]
	}
}

func imageMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}
