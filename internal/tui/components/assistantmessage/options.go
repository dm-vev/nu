package assistantmessage

// Options configures an assistant message block.
type Options struct {
	PaddingX int
	PaddingY int

	TextStyle     func(string) string
	HeadingStyle  func(string) string
	StrongStyle   func(string) string
	EmphasisStyle func(string) string
	CodeStyle     func(string) string

	ThinkingStyle  func(string) string
	ThinkingStrong func(string) string

	ToolPendingBg func(string) string
	ToolSuccessBg func(string) string
	ToolErrorBg   func(string) string
	ToolTitle     func(string) string
	ToolText      func(string) string
	ToolErrorText func(string) string
	ToolAdded     func(string) string
	ToolRemoved   func(string) string
	ToolContext   func(string) string
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
	if opts.ThinkingStyle == nil {
		opts.ThinkingStyle = opts.TextStyle
	}
	if opts.ThinkingStrong == nil {
		opts.ThinkingStrong = opts.ThinkingStyle
	}
	if opts.ToolTitle == nil {
		opts.ToolTitle = opts.StrongStyle
	}
	if opts.ToolText == nil {
		opts.ToolText = opts.TextStyle
	}
	if opts.ToolErrorText == nil {
		opts.ToolErrorText = opts.ToolText
	}
	if opts.ToolAdded == nil {
		opts.ToolAdded = opts.ToolText
	}
	if opts.ToolRemoved == nil {
		opts.ToolRemoved = opts.ToolText
	}
	if opts.ToolContext == nil {
		opts.ToolContext = opts.ToolText
	}
	return opts
}

func identity(value string) string {
	return value
}
