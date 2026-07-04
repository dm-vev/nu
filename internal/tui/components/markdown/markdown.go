package markdown

// Markdown displays a small Markdown subset with ANSI styling.
type Markdown struct {
	source string
	opts   Options
}

// New creates a Markdown component.
func New(source string, opts Options) *Markdown {
	return &Markdown{source: source, opts: normalizeOptions(opts)}
}

// SetText changes source Markdown.
func (m *Markdown) SetText(source string) {
	m.source = source
}

// Text returns raw Markdown source.
func (m *Markdown) Text() string {
	return m.source
}

// Invalidate exists for the component interface.
func (m *Markdown) Invalidate() {}
