package header

import (
	"strings"

	"nu/internal/tui/ansi"
)

// Render returns padded, wrapping header lines.
func (h *Header) Render(width int) []string {
	if width <= 0 {
		width = 1
	}
	contentWidth := width - h.opts.PaddingX*2
	if contentWidth < 1 {
		contentWidth = 1
	}
	lines := make([]string, 0)
	for i := 0; i < h.opts.PaddingY; i++ {
		lines = append(lines, strings.Repeat(" ", width))
	}
	left := strings.Repeat(" ", h.opts.PaddingX)
	right := strings.Repeat(" ", h.opts.PaddingX)
	for _, line := range ansi.WrapText(h.content(), contentWidth) {
		lines = append(lines, ansi.PadRight(left+line+right, width))
	}
	for i := 0; i < h.opts.PaddingY; i++ {
		lines = append(lines, strings.Repeat(" ", width))
	}
	return lines
}
