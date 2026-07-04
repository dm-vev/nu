package ansi

// ResetSuffix matches Pi's per-line reset.
const ResetSuffix = "\x1b[0m\x1b]8;;\x07"

// End returns the byte index after the escape sequence at start.
func End(text string, start int) int {
	if start+1 >= len(text) {
		return start + 1
	}
	switch text[start+1] {
	case '[':
		i := start + 2
		for i < len(text) {
			ch := text[i]
			i++
			if ch >= '@' && ch <= '~' {
				return i
			}
		}
	case ']', '_':
		i := start + 2
		for i < len(text) {
			if text[i] == 0x07 {
				return i + 1
			}
			if text[i] == 0x1b && i+1 < len(text) && text[i+1] == '\\' {
				return i + 2
			}
			i++
		}
	}
	return len(text)
}
