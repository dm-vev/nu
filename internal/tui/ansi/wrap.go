package ansi

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// WrapText wraps text to width while preserving ANSI sequences.
func WrapText(text string, width int) []string {
	if width <= 0 {
		width = 1
	}
	text = strings.ReplaceAll(text, "\t", "   ")
	paragraphs := strings.Split(text, "\n")
	lines := []string{}
	for _, paragraph := range paragraphs {
		lines = append(lines, wrapParagraph(paragraph, width)...)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func wrapParagraph(text string, width int) []string {
	if text == "" {
		return []string{""}
	}
	var lines []string
	var current strings.Builder
	currentWidth := 0
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			next := End(text, i)
			current.WriteString(text[i:next])
			i = next
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		w := RuneWidth(r)
		if currentWidth+w > width && current.Len() > 0 {
			line, rest := splitAtLastSpace(current.String())
			lines = append(lines, line)
			current.Reset()
			current.WriteString(rest)
			currentWidth = VisibleWidth(rest)
			if unicode.IsSpace(r) {
				i += size
				continue
			}
		}
		current.WriteRune(r)
		currentWidth += w
		i += size
	}
	lines = append(lines, strings.TrimRight(current.String(), " "))
	return lines
}

func splitAtLastSpace(text string) (string, string) {
	last := -1
	for i, r := range text {
		if unicode.IsSpace(r) {
			last = i
		}
	}
	if last <= 0 {
		return text, ""
	}
	return strings.TrimRight(text[:last], " "), strings.TrimLeft(text[last:], " ")
}
