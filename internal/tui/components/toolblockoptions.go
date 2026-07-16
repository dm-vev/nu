package components

// ToolBlockOptions configures a tool execution block.
type ToolBlockOptions struct {
	PaddingX int
	PaddingY int

	PendingBg func(string) string
	SuccessBg func(string) string
	ErrorBg   func(string) string

	TitleStyle   func(string) string
	TextStyle    func(string) string
	ErrorStyle   func(string) string
	AddedStyle   func(string) string
	RemovedStyle func(string) string
	ContextStyle func(string) string
}

func toolBlockNormalizeOptions(opts ToolBlockOptions) ToolBlockOptions {
	if opts.PaddingX < 0 {
		opts.PaddingX = 0
	}
	if opts.PaddingY < 0 {
		opts.PaddingY = 0
	}
	if opts.TitleStyle == nil {
		opts.TitleStyle = toolBlockIdentity
	}
	if opts.TextStyle == nil {
		opts.TextStyle = toolBlockIdentity
	}
	if opts.ErrorStyle == nil {
		opts.ErrorStyle = opts.TextStyle
	}
	if opts.AddedStyle == nil {
		opts.AddedStyle = opts.TextStyle
	}
	if opts.RemovedStyle == nil {
		opts.RemovedStyle = opts.TextStyle
	}
	if opts.ContextStyle == nil {
		opts.ContextStyle = opts.TextStyle
	}
	return opts
}

func toolBlockIdentity(value string) string {
	return value
}
