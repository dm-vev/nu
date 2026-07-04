package box

import "strings"

// Render renders children inside padded box lines.
func (b *Box) Render(width int) []string {
	if len(b.Children) == 0 {
		return nil
	}
	contentWidth := width - b.opts.PaddingX*2
	if contentWidth < 1 {
		contentWidth = 1
	}
	lines := []string{}
	for i := 0; i < b.opts.PaddingY; i++ {
		lines = append(lines, b.applyBackground("", width))
	}
	left := strings.Repeat(" ", b.opts.PaddingX)
	for _, child := range b.Children {
		for _, line := range child.Render(contentWidth) {
			lines = append(lines, b.applyBackground(left+line, width))
		}
	}
	for i := 0; i < b.opts.PaddingY; i++ {
		lines = append(lines, b.applyBackground("", width))
	}
	return lines
}
