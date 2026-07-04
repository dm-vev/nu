package text

// Text displays wrapped multi-line text.
type Text struct {
	text  string
	opts  Options
	cache cache
}

// New creates a text component.
func New(value string, opts Options) *Text {
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
