package text

type cache struct {
	text  string
	width int
	lines []string
}

func (t *Text) cached(width int) ([]string, bool) {
	if t.cache.lines == nil || t.cache.text != t.text || t.cache.width != width {
		return nil, false
	}
	return append([]string(nil), t.cache.lines...), true
}

func (t *Text) store(width int, lines []string) {
	t.cache = cache{text: t.text, width: width, lines: append([]string(nil), lines...)}
}

// Invalidate clears cached lines.
func (t *Text) Invalidate() {
	t.cache = cache{}
}
