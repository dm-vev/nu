package components

// Footer renders cwd, branch, context usage, provider, and model identity.
type Footer struct {
	opts FooterOptions
}

// NewFooter creates a footer component.
func NewFooter(opts FooterOptions) *Footer {
	return &Footer{opts: footerNormalizeOptions(opts)}
}

// SetOptions replaces footer data.
func (f *Footer) SetOptions(opts FooterOptions) {
	f.opts = footerNormalizeOptions(opts)
}

// Options returns the current footer data.
func (f *Footer) Options() FooterOptions {
	return f.opts
}

// Invalidate exists for the component interface.
func (f *Footer) Invalidate() {}
