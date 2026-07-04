package ansi

import (
	"unicode"
	"unicode/utf8"
)

// VisibleWidth returns terminal cell width after removing escape sequences.
func VisibleWidth(text string) int {
	width := 0
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			i = End(text, i)
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == '\t' {
			width += 3
		} else {
			width += RuneWidth(r)
		}
		i += size
	}
	return width
}

// RuneWidth approximates terminal cell width.
func RuneWidth(r rune) int {
	if r == 0 || unicode.IsControl(r) || unicode.Is(unicode.Mn, r) {
		return 0
	}
	if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) ||
		(r >= 0x1f000 && r <= 0x1faff) {
		return 2
	}
	return 1
}
