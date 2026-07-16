package editor

import (
	"strings"
	"unicode/utf8"
)

const (
	bracketedPasteStart = "\x1b[200~"
	bracketedPasteEnd   = "\x1b[201~"
)

// HandleInput updates editor state from raw key text.
func (e *Editor) HandleInput(data string) {
	if strings.HasPrefix(data, bracketedPasteStart) && strings.HasSuffix(data, bracketedPasteEnd) {
		// Pasted content is inserted as literal text so embedded newlines create editor lines.
		e.insert(strings.TrimSuffix(strings.TrimPrefix(data, bracketedPasteStart), bracketedPasteEnd))
		return
	}
	switch data {
	case "\r", "\n":
		e.submit()
	case "\x7f", "\b":
		e.backspace()
	case "\x1b[3~":
		e.forwardDelete()
	case "\x1b[D":
		e.move(-1)
	case "\x1b[C":
		e.move(1)
	default:
		if data == "" || data[0] < 0x20 {
			return
		}
		e.insert(data)
	}
}

func (e *Editor) insert(text string) {
	line := []rune(e.state.Lines[e.state.CursorLine])
	cursor := clampRuneIndex(e.state.CursorCol, len(line))
	before := string(line[:cursor])
	after := string(line[cursor:])
	text = strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(text, "\n")
	if len(parts) == 1 {
		e.state.Lines[e.state.CursorLine] = before + text + after
		e.state.CursorCol += utf8.RuneCountInString(text)
		e.changed()
		return
	}
	next := append([]string{}, e.state.Lines[:e.state.CursorLine]...)
	next = append(next, before+parts[0])
	next = append(next, parts[1:len(parts)-1]...)
	next = append(next, parts[len(parts)-1]+after)
	next = append(next, e.state.Lines[e.state.CursorLine+1:]...)
	e.state.Lines = next
	e.state.CursorLine += len(parts) - 1
	e.state.CursorCol = utf8.RuneCountInString(parts[len(parts)-1])
	e.changed()
}

func (e *Editor) backspace() {
	if e.state.CursorCol > 0 {
		line := []rune(e.state.Lines[e.state.CursorLine])
		cursor := clampRuneIndex(e.state.CursorCol, len(line))
		e.state.Lines[e.state.CursorLine] = string(line[:cursor-1]) + string(line[cursor:])
		e.state.CursorCol--
		e.changed()
		return
	}
	if e.state.CursorLine == 0 {
		return
	}
	prev := e.state.Lines[e.state.CursorLine-1]
	current := e.state.Lines[e.state.CursorLine]
	e.state.Lines[e.state.CursorLine-1] = prev + current
	e.state.Lines = append(e.state.Lines[:e.state.CursorLine], e.state.Lines[e.state.CursorLine+1:]...)
	e.state.CursorLine--
	e.state.CursorCol = utf8.RuneCountInString(prev)
	e.changed()
}

func (e *Editor) forwardDelete() {
	line := []rune(e.state.Lines[e.state.CursorLine])
	cursor := clampRuneIndex(e.state.CursorCol, len(line))
	if cursor < len(line) {
		e.state.Lines[e.state.CursorLine] = string(line[:cursor]) + string(line[cursor+1:])
		e.changed()
		return
	}
	if e.state.CursorLine >= len(e.state.Lines)-1 {
		return
	}
	e.state.Lines[e.state.CursorLine] += e.state.Lines[e.state.CursorLine+1]
	e.state.Lines = append(e.state.Lines[:e.state.CursorLine+1], e.state.Lines[e.state.CursorLine+2:]...)
	e.changed()
}

func (e *Editor) move(delta int) {
	e.state.CursorCol += delta
	if e.state.CursorCol < 0 {
		e.state.CursorCol = 0
	}
	lineLen := utf8.RuneCountInString(e.state.Lines[e.state.CursorLine])
	if e.state.CursorCol > lineLen {
		e.state.CursorCol = lineLen
	}
}

func clampRuneIndex(index int, length int) int {
	if index < 0 {
		return 0
	}
	if index > length {
		return length
	}
	return index
}

func (e *Editor) changed() {
	if e.onChange != nil {
		e.onChange(e.Text())
	}
}
