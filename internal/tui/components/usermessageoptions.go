package components

// UserMessageOptions configures a user message block.
type UserMessageOptions struct {
	PaddingX int
	PaddingY int

	TextStyle     func(string) string
	StrongStyle   func(string) string
	EmphasisStyle func(string) string
	CodeStyle     func(string) string
	Background    func(string) string
}

func userMessageNormalizeOptions(opts UserMessageOptions) UserMessageOptions {
	if opts.PaddingX < 0 {
		opts.PaddingX = 0
	}
	if opts.PaddingY < 0 {
		opts.PaddingY = 0
	}
	if opts.TextStyle == nil {
		opts.TextStyle = userMessageIdentity
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

func userMessageIdentity(value string) string {
	return value
}
