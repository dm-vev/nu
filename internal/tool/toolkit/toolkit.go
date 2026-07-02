package toolkit

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"nu/internal/agent"
)

// MutationMu serializes Phase 2 file mutations.
// ponytail: global mutation lock, per-path locks if write/edit throughput matters.
var MutationMu sync.Mutex

// DecodeArgs decodes raw tool-call JSON into out.
func DecodeArgs(raw string, out any) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return fmt.Errorf("decode tool args: %w", err)
	}
	return nil
}

// ResolveUnder resolves requested under cwd and rejects cwd escapes.
func ResolveUnder(cwd, requested string) (string, string, error) {
	if strings.TrimSpace(cwd) == "" {
		return "", "", errors.New("missing cwd")
	}
	if strings.TrimSpace(requested) == "" {
		return "", "", errors.New("missing path")
	}
	if filepath.IsAbs(requested) {
		return "", "", fmt.Errorf("path %q must be relative", requested)
	}
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return "", "", fmt.Errorf("resolve cwd %q: %w", cwd, err)
	}
	realCWD, err := filepath.EvalSymlinks(absCWD)
	if err != nil {
		return "", "", fmt.Errorf("resolve cwd %q: %w", cwd, err)
	}
	clean := filepath.Clean(requested)
	if clean == "." {
		return realCWD, ".", nil
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q escapes cwd", requested)
	}
	path := filepath.Join(realCWD, clean)
	realPath, err := resolveForContainment(path)
	if err != nil {
		return "", "", err
	}
	if !pathUnder(realCWD, realPath) {
		return "", "", fmt.Errorf("path %q escapes cwd", requested)
	}
	return realPath, filepath.ToSlash(clean), nil
}

func resolveForContainment(path string) (string, error) {
	var missing []string
	current := path
	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			for i := len(missing) - 1; i >= 0; i-- {
				real = filepath.Join(real, missing[i])
			}
			return real, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("resolve path %q: %w", path, err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("resolve path %q: %w", path, err)
		}
		// Missing leaf components are allowed for write/create tools, but the nearest
		// existing parent must still resolve under cwd.
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func pathUnder(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// Relative returns a slash-form path relative to cwd.
func Relative(cwd, path string) (string, error) {
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	return filepath.ToSlash(rel), nil
}

// TruncateString cuts value to maxBytes.
func TruncateString(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value, false
	}
	return value[:maxBytes], true
}

// JSONResult returns one JSON object as a tool result.
func JSONResult(value map[string]any) (agent.ToolResult, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("marshal tool result: %w", err)
	}
	return agent.ToolResult{Content: string(data)}, nil
}

// JSONListResult returns a JSON list result, dropping trailing values to fit maxBytes.
func JSONListResult(key string, values []string, maxBytes int) (agent.ToolResult, error) {
	if values == nil {
		values = []string{}
	}
	truncated := false
	for {
		result := map[string]any{key: values, "truncated": truncated}
		data, err := json.Marshal(result)
		if err != nil {
			return agent.ToolResult{}, fmt.Errorf("marshal tool result: %w", err)
		}
		if maxBytes <= 0 || len(data) <= maxBytes || len(values) == 0 {
			return agent.ToolResult{Content: string(data)}, nil
		}
		truncated = true
		values = values[:len(values)-1]
	}
}

// ImageMime recognizes image attachments supported by Phase 2.
func ImageMime(path string) string {
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

// PersistTempOutput stores full output when display output is truncated.
func PersistTempOutput(output string) (string, error) {
	file, err := os.CreateTemp("", "nu-bash-*.log")
	if err != nil {
		return "", fmt.Errorf("create temp output: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString(output); err != nil {
		return "", fmt.Errorf("write temp output %s: %w", file.Name(), err)
	}
	return file.Name(), nil
}

// LoadGitignore reads simple root .gitignore patterns.
func LoadGitignore(root string) []string {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return nil
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, filepath.ToSlash(line))
	}
	return patterns
}

// ShouldSkip applies Phase 2 ignore rules.
func ShouldSkip(rel string, isDir bool, patterns []string) bool {
	if rel == "." {
		return false
	}
	base := filepath.Base(rel)
	if base == ".git" && isDir {
		return true
	}
	for _, pattern := range patterns {
		dirPattern := strings.TrimSuffix(pattern, "/")
		if strings.HasSuffix(pattern, "/") && (rel == dirPattern || strings.HasPrefix(rel, dirPattern+"/")) {
			return true
		}
		if WildcardMatch(pattern, rel) || WildcardMatch(pattern, base) {
			return true
		}
		if rel == pattern || base == pattern || strings.HasPrefix(rel, pattern+"/") {
			return true
		}
	}
	return false
}

// GlobMatches checks a glob against a relative path or base name.
func GlobMatches(pattern, rel string) bool {
	if pattern == "" {
		return true
	}
	return WildcardMatch(pattern, rel) || WildcardMatch(pattern, filepath.Base(rel))
}

// WildcardMatch wraps filepath.Match with invalid-pattern-as-no-match behavior.
func WildcardMatch(pattern, value string) bool {
	ok, err := filepath.Match(pattern, value)
	return err == nil && ok
}

// DefaultPath returns fallback when value is empty.
func DefaultPath(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
