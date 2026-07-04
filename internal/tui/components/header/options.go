package header

// Options configures the startup header.
type Options struct {
	AppName string
	Version string

	Accent func(string) string
	Dim    func(string) string
	Muted  func(string) string

	HelpSeparator string
	PaddingX      int
	PaddingY      int
}

func normalizeOptions(opts Options) Options {
	if opts.AppName == "" {
		opts.AppName = "Nu"
	}
	if opts.Accent == nil {
		opts.Accent = identity
	}
	if opts.Dim == nil {
		opts.Dim = identity
	}
	if opts.Muted == nil {
		opts.Muted = identity
	}
	if opts.HelpSeparator == "" {
		opts.HelpSeparator = " · "
	}
	if opts.PaddingX < 0 {
		opts.PaddingX = 0
	}
	if opts.PaddingY < 0 {
		opts.PaddingY = 0
	}
	return opts
}

func identity(value string) string {
	return value
}
