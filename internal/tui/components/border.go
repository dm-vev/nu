package components

// Border renders a horizontal rule.
type Border struct {
	style func(string) string
}

// NewBorder creates a border component.
func NewBorder(style func(string) string) *Border {
	return &Border{style: style}
}
