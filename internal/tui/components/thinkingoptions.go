package components

// ThinkingOptions configures model-thinking rendering.
type ThinkingOptions struct {
	PaddingX int
	PaddingY int

	TextStyle     func(string) string
	StrongStyle   func(string) string
	EmphasisStyle func(string) string
	CodeStyle     func(string) string
}

func thinkingNormalizeOptions(opts ThinkingOptions) ThinkingOptions {
	if opts.PaddingX < 0 {
		opts.PaddingX = 0
	}
	if opts.PaddingY < 0 {
		opts.PaddingY = 0
	}
	if opts.TextStyle == nil {
		opts.TextStyle = thinkingIdentity
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

func thinkingIdentity(value string) string {
	return value
}
