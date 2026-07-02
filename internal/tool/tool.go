package tool

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"nu/internal/agent"
)

const defaultMaxOutputBytes = 16 * 1024

var mutationMu sync.Mutex

// Builtins returns the Phase 2 built-in tool set.
func Builtins(cwd string) map[string]agent.ToolFunc {
	return map[string]agent.ToolFunc{
		"read": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Read(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"write": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Write(ctx, cwd, call.Arguments)
		},
		"edit": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Edit(ctx, cwd, call.Arguments)
		},
		"bash": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Bash(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"grep": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Grep(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"find": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Find(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
		"ls": func(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error) {
			return Ls(ctx, cwd, call.Arguments, defaultMaxOutputBytes)
		},
	}
}

// Read reads one text or supported image file under cwd.
func Read(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("read %s: %w", rel, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("stat %s: %w", rel, err)
	}
	if info.IsDir() {
		return agent.ToolResult{}, fmt.Errorf("read %s: is a directory", rel)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("read %s: %w", rel, err)
	}
	if mimeType := imageMime(path); mimeType != "" {
		return jsonResult(map[string]any{
			"path":      rel,
			"mime_type": mimeType,
			"data":      base64.StdEncoding.EncodeToString(data),
		})
	}
	if args.Offset < 0 || args.Offset > len(data) {
		return agent.ToolResult{}, fmt.Errorf("read %s: invalid offset", rel)
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

// Write creates or overwrites one file under cwd.
func Write(ctx context.Context, cwd string, raw string) (agent.ToolResult, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}

	// ponytail: global mutation lock, per-path locks if write throughput matters.
	mutationMu.Lock()
	defer mutationMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return agent.ToolResult{}, fmt.Errorf("create parent for %s: %w", rel, err)
	}
	if err := os.WriteFile(path, []byte(args.Content), 0o644); err != nil {
		return agent.ToolResult{}, fmt.Errorf("write %s: %w", rel, err)
	}
	return jsonResult(map[string]any{"path": rel, "bytes": len(args.Content)})
}

// Edit applies exact replacements to one file under cwd.
func Edit(ctx context.Context, cwd string, raw string) (agent.ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Replacements []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"replacements"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if len(args.Replacements) == 0 {
		return agent.ToolResult{}, errors.New("edit: missing replacements")
	}
	path, rel, err := resolveUnder(cwd, args.Path)
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("edit %s: %w", rel, err)
	}

	// ponytail: global mutation lock, per-path locks if write throughput matters.
	mutationMu.Lock()
	defer mutationMu.Unlock()

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
	return jsonResult(map[string]any{"path": rel, "patch": patch.String()})
}

// Bash runs one shell command in cwd.
func Bash(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Command   string `json:"command"`
		TimeoutMS int    `json:"timeout_ms"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if strings.TrimSpace(args.Command) == "" {
		return agent.ToolResult{}, errors.New("bash: missing command")
	}
	runCtx := ctx
	cancel := func() {}
	if args.TimeoutMS > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(args.TimeoutMS)*time.Millisecond)
	}
	defer cancel()

	cmd := exec.CommandContext(runCtx, "sh", "-c", args.Command)
	cmd.Dir = cwd
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	timedOut := errors.Is(runCtx.Err(), context.DeadlineExceeded)
	exitCode := 0
	if err != nil {
		exitCode = 1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		if timedOut {
			exitCode = -1
		}
	}

	fullOutput := stdout.String() + stderr.String()
	output, truncated := truncateString(fullOutput, maxBytes)
	result := map[string]any{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
		"timed_out": timedOut,
		"output":    output,
		"truncated": truncated,
	}
	if truncated {
		path, err := persistTempOutput(fullOutput)
		if err != nil {
			return agent.ToolResult{}, err
		}
		result["full_output_path"] = path
	}
	return jsonResult(result)
}

// Grep searches files under cwd.
func Grep(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Pattern    string `json:"pattern"`
		Literal    bool   `json:"literal"`
		IgnoreCase bool   `json:"ignore_case"`
		Glob       string `json:"glob"`
		Root       string `json:"root"`
		Limit      int    `json:"limit"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if args.Pattern == "" {
		return agent.ToolResult{}, errors.New("grep: missing pattern")
	}
	if args.Limit <= 0 {
		args.Limit = 100
	}
	root, _, err := resolveUnder(cwd, defaultPath(args.Root, "."))
	if err != nil {
		return agent.ToolResult{}, err
	}
	matcher, err := newTextMatcher(args.Pattern, args.Literal, args.IgnoreCase)
	if err != nil {
		return agent.ToolResult{}, err
	}
	ignore := loadGitignore(root)
	matches := []string{}
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
		fileMatches, err := grepFile(path, rel, matcher, args.Limit-len(matches))
		if err != nil {
			return err
		}
		matches = append(matches, fileMatches...)
		if len(matches) >= args.Limit {
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("grep: %w", err)
	}
	sort.Strings(matches)
	return jsonListResult("matches", matches, maxBytes)
}

// Find finds files under cwd.
func Find(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Root  string `json:"root"`
		Glob  string `json:"glob"`
		Limit int    `json:"limit"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if args.Limit <= 0 {
		args.Limit = 100
	}
	root, _, err := resolveUnder(cwd, defaultPath(args.Root, "."))
	if err != nil {
		return agent.ToolResult{}, err
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
		return agent.ToolResult{}, fmt.Errorf("find: %w", err)
	}
	sort.Strings(paths)
	return jsonListResult("paths", paths, maxBytes)
}

// Ls lists one directory under cwd.
func Ls(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := decodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	path, rel, err := resolveUnder(cwd, defaultPath(args.Path, "."))
	if err != nil {
		return agent.ToolResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return agent.ToolResult{}, fmt.Errorf("ls %s: %w", rel, err)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("ls %s: %w", rel, err)
	}
	var names []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		names = append(names, name)
	}
	if len(entries) == 0 {
		info, err := os.Stat(path)
		if err != nil {
			return agent.ToolResult{}, fmt.Errorf("stat %s: %w", rel, err)
		}
		if !info.IsDir() {
			return agent.ToolResult{}, fmt.Errorf("ls %s: not a directory", rel)
		}
	}
	sort.Strings(names)
	return jsonListResult("entries", names, maxBytes)
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

func resolveUnder(cwd, requested string) (string, string, error) {
	if strings.TrimSpace(cwd) == "" {
		return "", "", errors.New("missing cwd")
	}
	if strings.TrimSpace(requested) == "" {
		return "", "", errors.New("missing path")
	}
	if filepath.IsAbs(requested) {
		return "", "", fmt.Errorf("path %q must be relative", requested)
	}
	clean := filepath.Clean(requested)
	if clean == "." {
		return filepath.Clean(cwd), ".", nil
	}
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q escapes cwd", requested)
	}
	return filepath.Join(cwd, clean), filepath.ToSlash(clean), nil
}

func relative(cwd, path string) (string, error) {
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return "", err
	}
	if rel == "." {
		return ".", nil
	}
	return filepath.ToSlash(rel), nil
}

func truncateString(value string, maxBytes int) (string, bool) {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value, false
	}
	return value[:maxBytes], true
}

func jsonResult(value map[string]any) (agent.ToolResult, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return agent.ToolResult{}, fmt.Errorf("marshal tool result: %w", err)
	}
	return agent.ToolResult{Content: string(data)}, nil
}

func jsonListResult(key string, values []string, maxBytes int) (agent.ToolResult, error) {
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

func imageMime(path string) string {
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

func persistTempOutput(output string) (string, error) {
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

type textMatcher func(string) bool

func newTextMatcher(pattern string, literal bool, ignoreCase bool) (textMatcher, error) {
	if ignoreCase {
		pattern = strings.ToLower(pattern)
	}
	if literal {
		return func(line string) bool {
			if ignoreCase {
				line = strings.ToLower(line)
			}
			return strings.Contains(line, pattern)
		}, nil
	}
	if ignoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile grep pattern: %w", err)
	}
	return re.MatchString, nil
}

func grepFile(path, rel string, matcher textMatcher, limit int) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", rel, err)
	}
	defer file.Close()
	var matches []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if matcher(line) {
			matches = append(matches, fmt.Sprintf("%s:%d:%s", rel, lineNumber, line))
			if len(matches) >= limit {
				return matches, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan %s: %w", rel, err)
	}
	return matches, nil
}

func loadGitignore(root string) []string {
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

func shouldSkip(rel string, isDir bool, patterns []string) bool {
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
		if wildcardMatch(pattern, rel) || wildcardMatch(pattern, base) {
			return true
		}
		if rel == pattern || base == pattern || strings.HasPrefix(rel, pattern+"/") {
			return true
		}
	}
	return false
}

func globMatches(pattern, rel string) bool {
	if pattern == "" {
		return true
	}
	return wildcardMatch(pattern, rel) || wildcardMatch(pattern, filepath.Base(rel))
}

func wildcardMatch(pattern, value string) bool {
	ok, err := filepath.Match(pattern, value)
	return err == nil && ok
}

func defaultPath(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
