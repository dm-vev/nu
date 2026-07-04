package ansi

import (
	"strings"
	"unicode/utf8"
)

// Strip removes escape sequences while leaving printable text.
func Strip(text string) string {
	var out strings.Builder
	for i := 0; i < len(text); {
		if text[i] == 0x1b {
			i = End(text, i)
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		out.WriteRune(r)
		i += size
	}
	return out.String()
}
