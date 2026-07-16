package components

import (
	"path/filepath"
	"strconv"
	"strings"
)

// FormatPath returns cwd shortened relative to home and annotated with branch.
func FormatPath(cwd string, home string, branch string) string {
	pathValue := footerFirstNonEmpty(cwd, ".")
	if home != "" {
		cleanHome := filepath.Clean(home)
		cleanCWD := filepath.Clean(pathValue)
		if cleanCWD == cleanHome {
			pathValue = "~"
		} else if rel, err := filepath.Rel(cleanHome, cleanCWD); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			pathValue = "~" + string(filepath.Separator) + rel
		}
	}
	if strings.TrimSpace(branch) != "" {
		pathValue += " (" + strings.TrimSpace(branch) + ")"
	}
	return pathValue
}

// FormatTokens returns a compact context token count.
func FormatTokens(count int) string {
	if count < 1000 {
		return strconv.Itoa(count)
	}
	if count < 10000 {
		return strconv.FormatFloat(float64(count)/1000, 'f', 1, 64) + "k"
	}
	if count < 1000000 {
		return strconv.Itoa(count/1000) + "k"
	}
	if count < 10000000 {
		return strconv.FormatFloat(float64(count)/1000000, 'f', 1, 64) + "M"
	}
	return strconv.Itoa(count/1000000) + "M"
}

func footerStatsLeft(used int, contextWindow int) string {
	percent := 0.0
	if used > 0 && contextWindow > 0 {
		percent = float64(used) * 100 / float64(contextWindow)
	}
	return strconv.FormatFloat(percent, 'f', 1, 64) + "%/" + FormatTokens(contextWindow) + " (auto)"
}

func footerModelRight(provider string, model string) string {
	return strings.Trim(strings.Join([]string{strings.TrimSpace(provider), strings.TrimSpace(model)}, "/"), "/")
}

func footerFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
