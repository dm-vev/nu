package components

import "strings"

// FillLines renders blank padded lines for the assigned height.
func (f *Fill) FillLines(width int, rows int) []string {
	if rows <= 0 {
		return nil
	}
	lines := make([]string, 0, rows)
	for i := 0; i < rows; i++ {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}
