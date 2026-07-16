package components

// Header renders Pi-style startup hints and onboarding text.
type Header struct {
	opts     HeaderOptions
	expanded bool
}

// NewHeader creates a header component.
func NewHeader(opts HeaderOptions) *Header {
	return &Header{opts: headerNormalizeOptions(opts)}
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
