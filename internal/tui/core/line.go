package core

import "nu/internal/tui/ansi"

// ResetLines appends Pi-compatible resets to non-image lines.
func ResetLines(lines []string) []string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = line + ansi.ResetSuffix
	}
	return out
}
