package text

// Options configures text rendering.
type Options struct {
	PaddingX int
	PaddingY int
	Bg       func(string) string
}
