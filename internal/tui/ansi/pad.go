package ansi

import "strings"

// PadRight pads visible cells to width.
func PadRight(text string, width int) string {
	visible := VisibleWidth(text)
	if visible >= width {
		return TruncateToWidth(text, width, "")
	}
	return text + strings.Repeat(" ", width-visible)
}
