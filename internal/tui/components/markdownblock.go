package components

import (
	"github.com/dm-vev/nu/internal/tui/ansi"
	"strings"
)

func markdownRenderBlocks(source string, opts MarkdownOptions) []string {
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	rawLines := strings.Split(source, "\n")

	lines := make([]string, 0, len(rawLines))
	inFence := false
	for i := 0; i < len(rawLines); i++ {
		line := strings.ReplaceAll(rawLines[i], "\t", "   ")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			lines = append(lines, opts.CodeStyle(line))
			continue
		}
		if trimmed == "" {
			lines = append(lines, "")
			continue
		}
		if table, next, ok := markdownParseTable(rawLines, i, opts); ok {
			lines = append(lines, table...)
			i = next
			continue
		}
		if heading, ok := markdownParseHeading(trimmed); ok {
			lines = append(lines, opts.HeadingStyle(heading))
			continue
		}
		if quote, ok := markdownParseQuote(trimmed); ok {
			lines = append(lines, opts.QuoteStyle("│ ")+markdownRenderInline(quote, opts))
			continue
		}
		if marker, body, ok := markdownParseList(trimmed); ok {
			lines = append(lines, opts.BulletStyle(marker)+" "+markdownRenderInline(body, opts))
			continue
		}
		lines = append(lines, markdownRenderInline(line, opts))
	}
	return lines
}

func markdownParseTable(rawLines []string, start int, opts MarkdownOptions) ([]string, int, bool) {
	if start+1 >= len(rawLines) || !markdownIsTableSeparator(rawLines[start+1]) {
		return nil, start, false
	}
	rows := [][]string{markdownSplitTableRow(rawLines[start])}
	if len(rows[0]) == 0 {
		return nil, start, false
	}
	i := start + 2
	for i < len(rawLines) && strings.Contains(rawLines[i], "|") && strings.TrimSpace(rawLines[i]) != "" {
		row := markdownSplitTableRow(rawLines[i])
		if len(row) == 0 {
			break
		}
		rows = append(rows, row)
		i++
	}
	widths := markdownTableWidths(rows, opts)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, markdownRenderTableRow(row, widths, opts))
	}
	return lines, i - 1, true
}

func markdownSplitTableRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func markdownIsTableSeparator(line string) bool {
	cells := markdownSplitTableRow(line)
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		cell = strings.Trim(cell, " :-")
		if cell != "" {
			return false
		}
	}
	return strings.Contains(line, "-") && strings.Contains(line, "|")
}

func markdownTableWidths(rows [][]string, opts MarkdownOptions) []int {
	widths := make([]int, 0)
	for _, row := range rows {
		for len(widths) < len(row) {
			widths = append(widths, 0)
		}
		for i, cell := range row {
			widths[i] = max(widths[i], ansi.VisibleWidth(markdownRenderInline(cell, opts)))
		}
	}
	return widths
}

func markdownRenderTableRow(row []string, widths []int, opts MarkdownOptions) string {
	cells := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		cells[i] = ansi.PadRight(markdownRenderInline(cell, opts), widths[i])
	}
	return strings.Join(cells, "  ")
}

func markdownParseHeading(line string) (string, bool) {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return "", false
	}
	return strings.TrimSpace(line[level+1:]), true
}

func markdownParseQuote(line string) (string, bool) {
	if !strings.HasPrefix(line, ">") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, ">")), true
}

func markdownParseList(line string) (string, string, bool) {
	for _, marker := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(line, marker) {
			return strings.TrimSpace(marker), strings.TrimSpace(line[len(marker):]), true
		}
	}
	for i := 0; i < len(line); i++ {
		if line[i] < '0' || line[i] > '9' {
			if i > 0 && i+1 < len(line) && line[i] == '.' && line[i+1] == ' ' {
				return line[:i+1], strings.TrimSpace(line[i+2:]), true
			}
			return "", "", false
		}
	}
	return "", "", false
}
