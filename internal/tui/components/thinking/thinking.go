package thinking

// Thinking renders a model reasoning block.
type Thinking struct {
	text string
	opts Options
}

// New creates a thinking component.
func New(value string, opts Options) *Thinking {
	return &Thinking{text: value, opts: normalizeOptions(opts)}
}

// SetText replaces reasoning text.
func (t *Thinking) SetText(value string) {
	t.text = value
}

// Text returns raw reasoning text.
func (t *Thinking) Text() string {
	return t.text
}

// Invalidate exists for the component interface.
func (t *Thinking) Invalidate() {}
