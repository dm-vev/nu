package components

import (
	"fmt"
	"github.com/dm-vev/nu/internal/model"
	"github.com/dm-vev/nu/internal/tui/ansi"
	"strings"
)

// Render draws the active model selector above the editor.
func (m *ModelMenu) Render(width int) []string {
	if !m.visible {
		return nil
	}
	if width <= 0 {
		width = 1
	}

	lines := []string{m.border(width)}
	lines = append(lines, m.line(width, m.opts.Muted("Only showing models from configured providers. Use /login to add providers.")))
	lines = append(lines, m.line(width, m.opts.Text("Search: ")+m.opts.Accent(m.query)))
	lines = append(lines, m.renderItems(width)...)
	lines = append(lines, m.border(width))
	return lines
}

func (m *ModelMenu) renderItems(width int) []string {
	if len(m.filtered) == 0 {
		return []string{m.line(width, m.opts.Muted("  No matching models"))}
	}

	start, end := m.visibleRange()
	lines := make([]string, 0, end-start+3)
	for index := start; index < end; index++ {
		lines = append(lines, m.line(width, m.renderItem(m.filtered[index], index == m.selected)))
	}
	if start > 0 || end < len(m.filtered) {
		lines = append(lines, m.line(width, m.opts.Muted(fmt.Sprintf("  (%d/%d)", m.selected+1, len(m.filtered)))))
	}
	if selected, ok := m.Selected(); ok {
		lines = append(lines, m.line(width, ""))
		lines = append(lines, m.line(width, m.opts.Muted("  Model Name: ")+m.opts.Text(modelMenuModelDisplayName(selected))))
	}
	return lines
}

func (m *ModelMenu) renderItem(entry model.Model, selected bool) string {
	prefix := "  "
	id := entry.ID
	if selected {
		prefix = "> "
		id = m.opts.Accent(id)
	}
	current := ""
	if m.isCurrent(entry) {
		current = m.opts.Success(" *")
	}
	return prefix + id + m.opts.Muted(" ["+entry.Provider+"]") + current
}

func (m *ModelMenu) visibleRange() (int, int) {
	maxVisible := m.opts.MaxVisible
	start := m.selected - maxVisible/2
	if start < 0 {
		start = 0
	}
	if start+maxVisible > len(m.filtered) {
		start = max(0, len(m.filtered)-maxVisible)
	}
	end := min(len(m.filtered), start+maxVisible)
	return start, end
}

func (m *ModelMenu) border(width int) string {
	return m.opts.Border(strings.Repeat("─", width))
}

func (m *ModelMenu) line(width int, text string) string {
	return ansi.PadRight(ansi.TruncateToWidth(text, width, ""), width)
}
