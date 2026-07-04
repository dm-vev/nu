package border

// Border renders a horizontal rule.
type Border struct {
	style func(string) string
}

// New creates a border component.
func New(style func(string) string) *Border {
	return &Border{style: style}
}
