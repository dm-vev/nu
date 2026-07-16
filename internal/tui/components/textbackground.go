package components

import "github.com/dm-vev/nu/internal/tui/ansi"

func (t *Text) applyBackground(line string, width int) string {
	line = ansi.PadRight(line, width)
	if t.opts.Bg == nil {
		return line
	}
	return t.opts.Bg(line)
}
