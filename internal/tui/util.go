package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func windowTitle(cwd string) string {
	base := filepath.Base(firstNonEmpty(cwd, "."))
	if base == "." || base == string(filepath.Separator) {
		return "Nu"
	}
	return "Nu - " + base
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
