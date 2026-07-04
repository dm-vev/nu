package ansi

import (
	"strings"
	"unicode/utf8"
)

// TruncateToWidth truncates text to width preserving escape sequences.
func TruncateToWidth(text string, width int, suffix string) string {
	if width <= 0 {
		return ""
	}
	if VisibleWidth(text) <= width {
		return text
	}
	limit := width - VisibleWidth(suffix)
	if limit < 0 {
		limit = 0
	}
	return SliceWidth(text, limit) + suffix
}

// SliceWidth returns the prefix fitting into width.
func SliceWidth(text string, width int) string {
	var out strings.Builder
	used := 0
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			next := End(text, i)
			out.WriteString(text[i:next])
			i = next
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		w := RuneWidth(r)
		if used+w > width {
			break
		}
		out.WriteRune(r)
		used += w
		i += size
	}
	return out.String()
}
