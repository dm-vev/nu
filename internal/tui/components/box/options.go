package box

// Options configures a box component.
type Options struct {
	PaddingX int
	PaddingY int
	Bg       func(string) string
}
