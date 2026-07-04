package markdown

import (
	"strings"

	"nu/internal/tui/ansi"
)

// Render formats Markdown blocks, wraps them, and applies component padding.
func (m *Markdown) Render(width int) []string {
	if strings.TrimSpace(m.source) == "" {
		return nil
	}
	contentWidth := width - m.opts.PaddingX*2
	if contentWidth < 1 {
		contentWidth = 1
	}

	blockLines := trimBlankEdges(renderBlocks(m.source, m.opts))
	if len(blockLines) == 0 {
		return nil
	}

	lines := make([]string, 0, len(blockLines)+m.opts.PaddingY*2)
	empty := strings.Repeat(" ", width)
	for i := 0; i < m.opts.PaddingY; i++ {
		lines = append(lines, empty)
	}

	left := strings.Repeat(" ", m.opts.PaddingX)
	right := strings.Repeat(" ", m.opts.PaddingX)
	for _, line := range blockLines {
		if line == "" {
			lines = append(lines, strings.Repeat(" ", width))
			continue
		}
		for _, wrapped := range ansi.WrapText(line, contentWidth) {
			lines = append(lines, ansi.PadRight(left+wrapped+right, width))
		}
	}

	for i := 0; i < m.opts.PaddingY; i++ {
		lines = append(lines, empty)
	}
	return lines
}

func trimBlankEdges(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(ansi.Strip(lines[start])) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(ansi.Strip(lines[end-1])) == "" {
		end--
	}
	return lines[start:end]
}
