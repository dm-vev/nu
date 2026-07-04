package header

// Header renders Pi-style startup hints and onboarding text.
type Header struct {
	opts     Options
	expanded bool
}

// New creates a header component.
func New(opts Options) *Header {
	return &Header{opts: normalizeOptions(opts)}
}

// SetExpanded switches between compact and full help.
func (h *Header) SetExpanded(expanded bool) {
	h.expanded = expanded
}

// Toggle flips the expansion state.
func (h *Header) Toggle() {
	h.expanded = !h.expanded
}

// Expanded reports the current expansion state.
func (h *Header) Expanded() bool {
	return h.expanded
}

// Invalidate exists for the component interface.
func (h *Header) Invalidate() {}
