package coding

import (
	"os"
	"path/filepath"
	"strings"
)

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
