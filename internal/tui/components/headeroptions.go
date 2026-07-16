package components

// HeaderOptions configures the startup header.
type HeaderOptions struct {
	AppName string
	Version string

	Accent func(string) string
	Dim    func(string) string
	Muted  func(string) string

	HelpSeparator string
	PaddingX      int
	PaddingY      int
}

func headerNormalizeOptions(opts HeaderOptions) HeaderOptions {
	if opts.AppName == "" {
		opts.AppName = "Nu"
	}
	if opts.Accent == nil {
		opts.Accent = headerIdentity
	}
	if opts.Dim == nil {
		opts.Dim = headerIdentity
	}
	if opts.Muted == nil {
		opts.Muted = headerIdentity
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

func headerIdentity(value string) string {
	return value
}
