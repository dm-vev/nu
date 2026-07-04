package modelmenu

// Options configures model selector rendering.
type Options struct {
	MaxVisible int
	Text       func(string) string
	Accent     func(string) string
	Muted      func(string) string
	Success    func(string) string
	Border     func(string) string
	Error      func(string) string
}

func normalizeOptions(opts Options) Options {
	if opts.MaxVisible <= 0 {
		opts.MaxVisible = 10
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
	if opts.Success == nil {
		opts.Success = opts.Accent
	}
	if opts.Border == nil {
		opts.Border = opts.Accent
	}
	if opts.Error == nil {
		opts.Error = opts.Text
	}
	return opts
}

func identity(value string) string {
	return value
}
