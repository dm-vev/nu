package components

import "nu/internal/tui/core"

// Box applies padding and background around child components.
type Box struct {
	core.Container
	opts BoxOptions
}

// NewBox creates a box.
func NewBox(opts BoxOptions) *Box {
	return &Box{opts: opts}
}
