package grep

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"nu/internal/agent"
	"nu/internal/tool/toolkit"
)

type textMatcher func(string) bool

const (
	maxScanTokenBytes = 1024 * 1024
	maxMatchLineBytes = 4096
)

// Run searches files under cwd.
func Run(ctx context.Context, cwd string, raw string, maxBytes int) (agent.ToolResult, error) {
	var args struct {
		Pattern    string `json:"pattern"`
		Literal    bool   `json:"literal"`
		IgnoreCase bool   `json:"ignore_case"`
		Glob       string `json:"glob"`
		Root       string `json:"root"`
		Limit      int    `json:"limit"`
	}
	if err := toolkit.DecodeArgs(raw, &args); err != nil {
		return agent.ToolResult{}, err
	}
	if args.Pattern == "" {
		return agent.ToolResult{}, errors.New("grep: missing pattern")
	}
	if args.Limit <= 0 {
		args.Limit = 100
	}
	root, _, err := toolkit.ResolveUnder(cwd, toolkit.DefaultPath(args.Root, "."))
	if err != nil {
		return agent.ToolResult{}, err
	}
	matcher, err := newTextMatcher(args.Pattern, args.Literal, args.IgnoreCase)
	if err != nil {
		return agent.ToolResult{}, err
	}
	ignore := toolkit.LoadGitignore(root)
	matches := []string{}
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
	return toolkit.JSONListResult("matches", matches, maxBytes)
}

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
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenBytes)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if matcher(line) {
			// Keep one very long matching line from eating the entire tool result.
			display, truncated := toolkit.TruncateString(line, maxMatchLineBytes)
			if truncated {
				display += "...[truncated]"
			}
			matches = append(matches, fmt.Sprintf("%s:%d:%s", rel, lineNumber, display))
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
