package engine

import (
	"strings"

	"github.com/dm-vev/nu/internal/tui/ansi"
	"github.com/dm-vev/nu/internal/tui/core"
)

// RenderNow renders the current component tree.
func (t *TUI) RenderNow() error {
	width, rows := t.terminal.Size()
	lines := t.renderAnchored(width, rows)
	targetRows := max(rows, t.opts.MinRenderRows)
	if len(lines) > targetRows {
		maxOffset := len(lines) - targetRows
		if t.scrollOffset > maxOffset {
			t.scrollOffset = maxOffset
		}
		start := max(0, maxOffset-t.scrollOffset)
		lines = lines[start : start+targetRows]
	} else {
		t.scrollOffset = 0
	}
	cursor := extractCursor(lines, targetRows)
	lines = core.ResetLines(lines)
	if len(lines) < targetRows {
		for len(lines) < targetRows {
			lines = append(lines, strings.Repeat(" ", width)+ansi.ResetSuffix)
		}
	}
	if len(t.previousLines) == 0 {
		return t.fullRender(lines, width, rows, true, cursor)
	}
	if t.previousWidth != width || t.previousRows != rows {
		return t.fullRender(lines, width, rows, true, cursor)
	}
	return t.diffRender(lines, width, rows, cursor)
}

func (t *TUI) renderAnchored(width int, rows int) []string {
	type chunk struct {
		filler core.Filler
		lines  []string
	}
	chunks := make([]chunk, 0, len(t.Children))
	fixedRows := 0
	fillers := 0
	for _, child := range t.Children {
		if filler, ok := child.(core.Filler); ok {
			chunks = append(chunks, chunk{filler: filler})
			fillers++
			continue
		}
		lines := child.Render(width)
		chunks = append(chunks, chunk{lines: lines})
		fixedRows += len(lines)
	}

	// Filler rows anchor later components, such as editor/footer, to the terminal bottom.
	remaining := max(0, rows-fixedRows)
	lines := []string{}
	for _, item := range chunks {
		if item.filler == nil {
			lines = append(lines, item.lines...)
			continue
		}
		share := 0
		if fillers > 0 {
			share = remaining / fillers
			remaining -= share
			fillers--
		}
		lines = append(lines, item.filler.FillLines(width, share)...)
	}
	return lines
}
