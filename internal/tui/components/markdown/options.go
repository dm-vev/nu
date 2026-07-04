package markdown

// Options configures Markdown rendering.
type Options struct {
	PaddingX int
	PaddingY int

	TextStyle     func(string) string
	HeadingStyle  func(string) string
	StrongStyle   func(string) string
	EmphasisStyle func(string) string
	CodeStyle     func(string) string
	QuoteStyle    func(string) string
	BulletStyle   func(string) string
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
	if opts.HeadingStyle == nil {
		opts.HeadingStyle = opts.TextStyle
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
	if opts.QuoteStyle == nil {
		opts.QuoteStyle = opts.TextStyle
	}
	if opts.BulletStyle == nil {
		opts.BulletStyle = opts.TextStyle
	}
	return opts
}

func identity(value string) string {
	return value
}
