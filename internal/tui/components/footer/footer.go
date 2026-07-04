package footer

// Footer renders cwd, branch, context usage, provider, and model identity.
type Footer struct {
	opts Options
}

// New creates a footer component.
func New(opts Options) *Footer {
	return &Footer{opts: normalizeOptions(opts)}
}

// SetOptions replaces footer data.
func (f *Footer) SetOptions(opts Options) {
	f.opts = normalizeOptions(opts)
}

// Options returns the current footer data.
func (f *Footer) Options() Options {
	return f.opts
}

// Invalidate exists for the component interface.
func (f *Footer) Invalidate() {}
