package engine

import (
	"fmt"
	"strings"

	"github.com/dm-vev/nu/internal/tui/terminal"
)

func (t *TUI) fullRender(lines []string, width int, rows int, clear bool, cursor cursorPosition) error {
	var buffer strings.Builder
	buffer.WriteString(terminal.SyncStart)
	if clear || t.opts.InitialClear {
		buffer.WriteString("\x1b[2J\x1b[H\x1b[3J")
	}
	for i, line := range lines {
		if i > 0 {
			buffer.WriteString("\r\n")
		}
		buffer.WriteString(line)
	}
	buffer.WriteString(terminal.SyncEnd)
	if err := t.terminal.Write(buffer.String()); err != nil {
		return err
	}
	t.previousLines = append([]string(nil), lines...)
	t.previousWidth = width
	t.previousRows = rows
	t.cursorRow = max(0, len(lines)-1)
	t.cursorCol = 0
	return t.positionCursor(cursor)
}

func (t *TUI) diffRender(lines []string, width int, rows int, cursor cursorPosition) error {
	first, last := changedRange(t.previousLines, lines)
	if first < 0 {
		return t.positionCursor(cursor)
	}
	var buffer strings.Builder
	buffer.WriteString(terminal.SyncStart)
	for i := first; i <= last && i < len(lines); i++ {
		buffer.WriteString(fmt.Sprintf("\x1b[%d;1H", i+1))
		buffer.WriteString("\x1b[2K")
		buffer.WriteString(lines[i])
	}
	if len(t.previousLines) > len(lines) {
		for i := len(lines); i < len(t.previousLines); i++ {
			buffer.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", i+1))
		}
	}
	buffer.WriteString(terminal.SyncEnd)
	if err := t.terminal.Write(buffer.String()); err != nil {
		return err
	}
	t.previousLines = append([]string(nil), lines...)
	t.previousWidth = width
	t.previousRows = rows
	t.cursorRow = min(last, max(0, len(lines)-1))
	t.cursorCol = 0
	return t.positionCursor(cursor)
}

func changedRange(oldLines []string, newLines []string) (int, int) {
	maxLines := max(len(oldLines), len(newLines))
	first := -1
	last := -1
	for i := 0; i < maxLines; i++ {
		oldLine := ""
		newLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}
		if oldLine == newLine {
			continue
		}
		if first < 0 {
			first = i
		}
		last = i
	}
	return first, last
}
