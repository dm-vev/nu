package usermessage

// Options configures a user message block.
type Options struct {
	PaddingX int
	PaddingY int

	TextStyle     func(string) string
	StrongStyle   func(string) string
	EmphasisStyle func(string) string
	CodeStyle     func(string) string
	Background    func(string) string
}

func normalizeOptions(opts Options) Options {
	if opts.PaddingX < 0 {
		opts.PaddingX = 0
	}
	if opts.PaddingY < 0 {
		opts.PaddingY = 0
	}
	if opts.TextStyle == nil {
		opts.TextStyle = identity
	}
	if opts.StrongStyle == nil {
		opts.StrongStyle = opts.TextStyle
	}
	if opts.EmphasisStyle == nil {
		opts.EmphasisStyle = opts.TextStyle
	}
	if opts.CodeStyle == nil {
		opts.CodeStyle = opts.TextStyle
	}
	return opts
}

func identity(value string) string {
	return value
}
