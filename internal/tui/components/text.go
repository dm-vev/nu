package components

// Text displays wrapped multi-line text.
type Text struct {
	text  string
	opts  TextOptions
	cache textCache
}

// NewText creates a text component.
func NewText(value string, opts TextOptions) *Text {
	return &Text{text: value, opts: opts}
}

// SetText changes text and clears cached lines.
func (t *Text) SetText(value string) {
	if t.text == value {
		return
	}
	t.text = value
	t.Invalidate()
}

// Text returns the current text.
func (t *Text) Text() string {
	return t.text
}
