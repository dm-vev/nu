package components

// Thinking renders a model reasoning block.
type Thinking struct {
	text string
	opts ThinkingOptions
}

// NewThinking creates a thinking component.
func NewThinking(value string, opts ThinkingOptions) *Thinking {
	return &Thinking{text: value, opts: thinkingNormalizeOptions(opts)}
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
