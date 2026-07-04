package commandmenu

import (
	"strings"

	"nu/internal/tui/ansi"
)

// Render draws the current command suggestions.
func (m *Menu) Render(width int) []string {
	if m.prefix == "" && len(m.matches) == 0 {
		return nil
	}
	if width <= 0 {
		width = 1
	}
	if len(m.matches) == 0 {
		return []string{ansi.PadRight(m.opts.Muted("  No matching commands"), width)}
	}
	lines := make([]string, 0, len(m.matches))
	for i, command := range m.matches {
		name := m.opts.Accent("/" + command.Name)
		desc := ""
		if command.Description != "" {
			desc = "  " + m.opts.Muted(command.Description)
		}
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		lines = append(lines, ansi.PadRight(prefix+name+desc, width))
	}
	return lines
}

func menuPrefix(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "/") || strings.Contains(text, " ") {
		return "", false
	}
	return text, true
}
