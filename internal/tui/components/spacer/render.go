package spacer

import "strings"

// Render returns blank padded lines.
func (s *Spacer) Render(width int) []string {
	lines := make([]string, 0, s.lines)
	for i := 0; i < s.lines; i++ {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}
