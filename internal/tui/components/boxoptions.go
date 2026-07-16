package components

// BoxOptions configures a box component.
type BoxOptions struct {
	PaddingX int
	PaddingY int
	Bg       func(string) string
}
