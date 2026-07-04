package editor

import (
	"strings"

	"nu/internal/tui/ansi"
	"nu/internal/tui/core"
)

// Render renders a bordered, wrapping editor.
func (e *Editor) Render(width int) []string {
	borderRune := e.borderRune
	if borderRune == 0 {
		borderRune = '─'
	}
	borderLine := strings.Repeat(string(borderRune), max(1, width))
	if e.border != nil {
		borderLine = e.border(borderLine)
	}
	contentWidth := max(1, width)
	lines := []string{borderLine}
	for lineIndex, logical := range e.state.Lines {
		if lineIndex == e.state.CursorLine {
			logical = insertMarker(logical, e.state.CursorCol)
		}
		wrapped := ansi.WrapText(logical, contentWidth)
		for _, visual := range wrapped {
			rendered := visual
			if e.textStyle != nil {
				rendered = e.textStyle(rendered)
			}
			lines = append(lines, ansi.PadRight(rendered, width))
		}
	}
	lines = append(lines, borderLine)
	return lines
}

func insertMarker(text string, col int) string {
	runes := []rune(text)
	if col < 0 {
		col = 0
	}
	if col > len(runes) {
		col = len(runes)
	}
	return string(runes[:col]) + core.CursorMarker + string(runes[col:])
}
