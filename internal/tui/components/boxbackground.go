package components

import "nu/internal/tui/ansi"

func (b *Box) applyBackground(line string, width int) string {
	line = ansi.PadRight(line, width)
	if b.opts.Bg == nil {
		return line
	}
	return b.opts.Bg(line)
}
