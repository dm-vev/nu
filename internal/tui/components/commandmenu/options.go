package commandmenu

// Options configures command menu rendering.
type Options struct {
	MaxItems int
	Text     func(string) string
	Accent   func(string) string
	Muted    func(string) string
}

func normalizeOptions(opts Options) Options {
	if opts.MaxItems <= 0 {
		opts.MaxItems = 8
	}
	if opts.Text == nil {
		opts.Text = identity
	}
	if opts.Accent == nil {
		opts.Accent = opts.Text
	}
	if opts.Muted == nil {
		opts.Muted = opts.Text
	}
	return opts
}

func identity(value string) string {
	return value
}
