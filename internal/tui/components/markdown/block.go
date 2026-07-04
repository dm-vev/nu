package markdown

import (
	"strings"

	"nu/internal/tui/ansi"
)

func renderBlocks(source string, opts Options) []string {
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
		if table, next, ok := parseTable(rawLines, i, opts); ok {
			lines = append(lines, table...)
			i = next
			continue
		}
		if heading, ok := parseHeading(trimmed); ok {
			lines = append(lines, opts.HeadingStyle(heading))
			continue
		}
		if quote, ok := parseQuote(trimmed); ok {
			lines = append(lines, opts.QuoteStyle("│ ")+renderInline(quote, opts))
			continue
		}
		if marker, body, ok := parseList(trimmed); ok {
			lines = append(lines, opts.BulletStyle(marker)+" "+renderInline(body, opts))
			continue
		}
		lines = append(lines, renderInline(line, opts))
	}
	return lines
}

func parseTable(rawLines []string, start int, opts Options) ([]string, int, bool) {
	if start+1 >= len(rawLines) || !isTableSeparator(rawLines[start+1]) {
		return nil, start, false
	}
	rows := [][]string{splitTableRow(rawLines[start])}
	if len(rows[0]) == 0 {
		return nil, start, false
	}
	i := start + 2
	for i < len(rawLines) && strings.Contains(rawLines[i], "|") && strings.TrimSpace(rawLines[i]) != "" {
		row := splitTableRow(rawLines[i])
		if len(row) == 0 {
			break
		}
		rows = append(rows, row)
		i++
	}
	widths := tableWidths(rows, opts)
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, renderTableRow(row, widths, opts))
	}
	return lines, i - 1, true
}

func splitTableRow(line string) []string {
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

func isTableSeparator(line string) bool {
	cells := splitTableRow(line)
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

func tableWidths(rows [][]string, opts Options) []int {
	widths := make([]int, 0)
	for _, row := range rows {
		for len(widths) < len(row) {
			widths = append(widths, 0)
		}
		for i, cell := range row {
			widths[i] = max(widths[i], ansi.VisibleWidth(renderInline(cell, opts)))
		}
	}
	return widths
}

func renderTableRow(row []string, widths []int, opts Options) string {
	cells := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		cells[i] = ansi.PadRight(renderInline(cell, opts), widths[i])
	}
	return strings.Join(cells, "  ")
}

func parseHeading(line string) (string, bool) {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return "", false
	}
	return strings.TrimSpace(line[level+1:]), true
}

func parseQuote(line string) (string, bool) {
	if !strings.HasPrefix(line, ">") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, ">")), true
}

func parseList(line string) (string, string, bool) {
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
