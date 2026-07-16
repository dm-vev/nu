package components

import (
	"github.com/dm-vev/nu/internal/tui/ansi"
	"strings"
)

// Render always reserves the status row directly above the editor.
func (s *Status) Render(width int) []string {
	if width <= 0 {
		width = 1
	}
	if strings.TrimSpace(s.text) == "" {
		return []string{strings.Repeat(" ", width)}
	}
	label := s.frames[s.frame%len(s.frames)] + " " + s.text
	if s.alert {
		return []string{ansi.PadRight(" "+statusAlertStyle(label, s.frame), width)}
	}
	return []string{ansi.PadRight(" "+s.style(label), width)}
}

func statusAlertStyle(value string, frame int) string {
	colors := []string{
		"\x1b[38;5;222m",
		"\x1b[38;5;214m",
		"\x1b[38;5;208m",
		"\x1b[38;5;203m",
		"\x1b[38;5;167m",
	}
	return colors[frame%len(colors)] + value + ansi.DefaultFG
}
