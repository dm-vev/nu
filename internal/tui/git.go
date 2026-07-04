package tui

import (
	"os"
	"path/filepath"
	"strings"
)

func currentGitBranch(cwd string) string {
	dir := firstNonEmpty(cwd, ".")
	for {
		data, err := os.ReadFile(filepath.Join(dir, ".git", "HEAD"))
		if err == nil {
			return strings.TrimPrefix(strings.TrimSpace(string(data)), "ref: refs/heads/")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
