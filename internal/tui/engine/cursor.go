package engine

import (
	"fmt"
	"strings"

	"github.com/dm-vev/nu/internal/tui/ansi"
	"github.com/dm-vev/nu/internal/tui/core"
)

type cursorPosition struct {
	row int
	col int
	ok  bool
}

func extractCursor(lines []string, rows int) cursorPosition {
	viewportTop := max(0, len(lines)-rows)
	for row := len(lines) - 1; row >= viewportTop; row-- {
		idx := strings.Index(lines[row], core.CursorMarker)
		if idx < 0 {
			continue
		}
		col := ansi.VisibleWidth(lines[row][:idx])
		lines[row] = lines[row][:idx] + lines[row][idx+len(core.CursorMarker):]
		return cursorPosition{row: row, col: col + 1, ok: true}
	}
	return cursorPosition{}
}

func (t *TUI) positionCursor(cursor cursorPosition) error {
	if !cursor.ok {
		return t.terminal.HideCursor()
	}
	diff := cursor.row - t.cursorRow
	if err := t.terminal.MoveBy(diff); err != nil {
		return err
	}
	if err := t.terminal.Write(fmt.Sprintf("\x1b[%dG", max(1, cursor.col))); err != nil {
		return err
	}
	t.cursorRow = cursor.row
	t.cursorCol = cursor.col
	if t.opts.ShowCursor {
		return t.terminal.ShowCursor()
	}
	return t.terminal.HideCursor()
}
