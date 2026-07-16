package components

// Markdown displays a small Markdown subset with ANSI styling.
type Markdown struct {
	source string
	opts   MarkdownOptions
}

// NewMarkdown creates a Markdown component.
func NewMarkdown(source string, opts MarkdownOptions) *Markdown {
	return &Markdown{source: source, opts: markdownNormalizeOptions(opts)}
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
