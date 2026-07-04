package footer

import (
	"strings"

	"nu/internal/tui/ansi"
)

// Render produces cwd and stats/model lines.
func (f *Footer) Render(width int) []string {
	if width <= 0 {
		width = 1
	}
	pathLine := alignStats(
		FormatPath(f.opts.CWD, f.opts.Home, f.opts.Branch),
		f.opts.Notice,
		width,
	)
	stats := statsLeft(f.opts.Used, f.opts.Context)
	right := modelRight(f.opts.Provider, f.opts.Model)
	statsLine := alignStats(stats, right, width)
	return []string{
		ansi.PadRight(stylePathLine(pathLine, f.opts), width),
		ansi.PadRight(f.opts.Dim(statsLine), width),
	}
}

func stylePathLine(line string, opts Options) string {
	if opts.Notice == "" {
		return opts.Dim(ansi.TruncateToWidth(line, ansi.VisibleWidth(line), opts.Dim("...")))
	}
	index := strings.LastIndex(line, opts.Notice)
	if index < 0 {
		return opts.Dim(line)
	}
	return opts.Dim(line[:index]) + opts.NoticeStyle(line[index:])
}

func alignStats(left string, right string, width int) string {
	if strings.TrimSpace(right) == "" {
		return ansi.TruncateToWidth(left, width, "")
	}
	leftWidth := ansi.VisibleWidth(left)
	rightWidth := ansi.VisibleWidth(right)
	if leftWidth+2+rightWidth <= width {
		return left + strings.Repeat(" ", width-leftWidth-rightWidth) + right
	}
	availableRight := width - leftWidth - 2
	if availableRight <= 0 {
		return ansi.TruncateToWidth(left, width, "")
	}
	return left + strings.Repeat(" ", 2) + ansi.TruncateToWidth(right, availableRight, "")
}
