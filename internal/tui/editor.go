package tui

// Editor owns the interactive input buffer.
type Editor struct {
	buffer        []rune
	cursor        int
	undo          []editorUndo
	lastSubmitted string
}

type editorUndo struct {
	buffer []rune
	cursor int
}

// EditorSnapshot is a read-only view of editor state.
type EditorSnapshot struct {
	Text          string
	Cursor        int
	LastSubmitted string
}

// NewEditor creates an empty editor.
func NewEditor() *Editor {
	return &Editor{}
}

// Insert inserts text at the current cursor.
func (e *Editor) Insert(text string) {
	if text == "" {
		return
	}
	e.saveUndo()
	runes := []rune(text)
	next := make([]rune, 0, len(e.buffer)+len(runes))
	next = append(next, e.buffer[:e.cursor]...)
	next = append(next, runes...)
	next = append(next, e.buffer[e.cursor:]...)
	e.buffer = next
	e.cursor += len(runes)
}

// Backspace removes the rune before the cursor.
func (e *Editor) Backspace() {
	if e.cursor == 0 {
		return
	}
	e.saveUndo()
	e.buffer = append(e.buffer[:e.cursor-1], e.buffer[e.cursor:]...)
	e.cursor--
}

// Move moves the cursor by delta runes.
func (e *Editor) Move(delta int) {
	e.cursor += delta
	if e.cursor < 0 {
		e.cursor = 0
	}
	if e.cursor > len(e.buffer) {
		e.cursor = len(e.buffer)
	}
}

// Undo restores the previous buffer and cursor snapshot.
func (e *Editor) Undo() bool {
	if len(e.undo) == 0 {
		return false
	}
	last := e.undo[len(e.undo)-1]
	e.undo = e.undo[:len(e.undo)-1]
	e.buffer = append([]rune(nil), last.buffer...)
	e.cursor = last.cursor
	return true
}

// Submit captures the current text and clears the input buffer.
func (e *Editor) Submit() string {
	text := string(e.buffer)
	// The submitted value is captured before clearing so callers get exact text.
	e.lastSubmitted = text
	e.buffer = nil
	e.cursor = 0
	e.undo = nil
	return text
}

// Snapshot returns a copy of visible editor state.
func (e *Editor) Snapshot() EditorSnapshot {
	return EditorSnapshot{
		Text:          string(e.buffer),
		Cursor:        e.cursor,
		LastSubmitted: e.lastSubmitted,
	}
}

func (e *Editor) saveUndo() {
	e.undo = append(e.undo, editorUndo{
		buffer: append([]rune(nil), e.buffer...),
		cursor: e.cursor,
	})
}
