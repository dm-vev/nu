package components

// ModelMenuOptions configures model selector rendering.
type ModelMenuOptions struct {
	MaxVisible int
	Text       func(string) string
	Accent     func(string) string
	Muted      func(string) string
	Success    func(string) string
	Border     func(string) string
	Error      func(string) string
}

func modelMenuNormalizeOptions(opts ModelMenuOptions) ModelMenuOptions {
	if opts.MaxVisible <= 0 {
		opts.MaxVisible = 10
	}
	if opts.Text == nil {
		opts.Text = modelMenuIdentity
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

func modelMenuIdentity(value string) string {
	return value
}
