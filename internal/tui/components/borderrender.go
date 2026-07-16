package components

import "strings"

// Render returns one horizontal border line.
func (b *Border) Render(width int) []string {
	line := strings.Repeat("─", max(1, width))
	if b.style != nil {
		line = b.style(line)
	}
	return []string{line}
}
