package components

// TextOptions configures text rendering.
type TextOptions struct {
	PaddingX int
	PaddingY int
	Bg       func(string) string
}
