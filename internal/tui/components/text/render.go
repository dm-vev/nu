package text

import (
	"strings"

	"nu/internal/tui/ansi"
)

// Render wraps text, applies margins, and pads to width.
func (t *Text) Render(width int) []string {
	if lines, ok := t.cached(width); ok {
		return lines
	}
	if strings.TrimSpace(t.text) == "" {
		return nil
	}
	contentWidth := width - t.opts.PaddingX*2
	if contentWidth < 1 {
		contentWidth = 1
	}
	lines := []string{}
	empty := strings.Repeat(" ", width)
	for i := 0; i < t.opts.PaddingY; i++ {
		lines = append(lines, t.applyBackground(empty, width))
	}
	left := strings.Repeat(" ", t.opts.PaddingX)
	right := strings.Repeat(" ", t.opts.PaddingX)
	for _, line := range ansi.WrapText(t.text, contentWidth) {
		lines = append(lines, t.applyBackground(left+line+right, width))
	}
	for i := 0; i < t.opts.PaddingY; i++ {
		lines = append(lines, t.applyBackground(empty, width))
	}
	t.store(width, lines)
	return lines
}
