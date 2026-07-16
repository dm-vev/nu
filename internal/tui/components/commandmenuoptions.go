package components

// CommandMenuOptions configures command menu rendering.
type CommandMenuOptions struct {
	MaxItems int
	Text     func(string) string
	Accent   func(string) string
	Muted    func(string) string
}

func commandMenuNormalizeOptions(opts CommandMenuOptions) CommandMenuOptions {
	if opts.MaxItems <= 0 {
		opts.MaxItems = 8
	}
	if opts.Text == nil {
		opts.Text = commandMenuIdentity
	}
	if opts.Accent == nil {
		opts.Accent = opts.Text
	}
	if opts.Muted == nil {
		opts.Muted = opts.Text
	}
	return opts
}

func commandMenuIdentity(value string) string {
	return value
}
