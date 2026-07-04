package editor

import "strings"

func (e *Editor) submit() {
	text := e.Text()
	e.state = initialState()
	if e.onSubmit != nil && strings.TrimSpace(text) != "" {
		e.onSubmit(text)
	}
	e.changed()
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n")
}
