package toolblock

import "strings"

func renderDiff(patch string, opts Options) string {
	rawLines := strings.Split(strings.ReplaceAll(patch, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		line = strings.ReplaceAll(line, "\t", "   ")
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			lines = append(lines, opts.ContextStyle(line))
		case strings.HasPrefix(line, "+"):
			lines = append(lines, opts.AddedStyle(line))
		case strings.HasPrefix(line, "-"):
			lines = append(lines, opts.RemovedStyle(line))
		case strings.HasPrefix(line, "@@"):
			lines = append(lines, opts.ContextStyle(line))
		default:
			lines = append(lines, opts.ContextStyle(line))
		}
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}
