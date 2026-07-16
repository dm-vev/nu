package components

import (
	"nu/internal/tui/ansi"
	"strings"
)

// Render produces cwd and stats/model lines.
func (f *Footer) Render(width int) []string {
	if width <= 0 {
		width = 1
	}
	pathLine := footerAlignStats(
		FormatPath(f.opts.CWD, f.opts.Home, f.opts.Branch),
		f.opts.Notice,
		width,
	)
	stats := footerStatsLeft(f.opts.Used, f.opts.Context)
	right := footerModelRight(f.opts.Provider, f.opts.Model)
	statsLine := footerAlignStats(stats, right, width)
	return []string{
		ansi.PadRight(footerStylePathLine(pathLine, f.opts), width),
		ansi.PadRight(f.opts.Dim(statsLine), width),
	}
}

func footerStylePathLine(line string, opts FooterOptions) string {
	if opts.Notice == "" {
		return opts.Dim(ansi.TruncateToWidth(line, ansi.VisibleWidth(line), opts.Dim("...")))
	}
	index := strings.LastIndex(line, opts.Notice)
	if index < 0 {
		return opts.Dim(line)
	}
	return opts.Dim(line[:index]) + opts.NoticeStyle(line[index:])
}

func footerAlignStats(left string, right string, width int) string {
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
