package editor

// Editor is a focused input component.
type Editor struct {
	state      State
	focused    bool
	onSubmit   func(string)
	onChange   func(string)
	border     func(string) string
	textStyle  func(string) string
	borderRune rune
}

// New creates an editor component.
func New() *Editor {
	return &Editor{state: initialState()}
}

// SetSubmitHandler sets the submit callback.
func (e *Editor) SetSubmitHandler(handler func(string)) {
	e.onSubmit = handler
}

// SetChangeHandler sets the change callback.
func (e *Editor) SetChangeHandler(handler func(string)) {
	e.onChange = handler
}

// SetStyles sets border and text styles.
func (e *Editor) SetStyles(border func(string) string, textStyle func(string) string) {
	e.border = border
	e.textStyle = textStyle
}

// SetBorderRune changes the editor border glyph.
func (e *Editor) SetBorderRune(borderRune rune) {
	e.borderRune = borderRune
}

// SetFocused updates focus state.
func (e *Editor) SetFocused(focused bool) {
	e.focused = focused
}

// Text returns editor text.
func (e *Editor) Text() string {
	return joinLines(e.state.Lines)
}

// Clear resets the editor buffer.
func (e *Editor) Clear() {
	e.state = initialState()
	e.changed()
}

// SetText replaces the editor buffer with one logical line.
func (e *Editor) SetText(value string) {
	e.state = State{Lines: []string{value}, CursorLine: 0, CursorCol: len([]rune(value))}
	e.changed()
}
