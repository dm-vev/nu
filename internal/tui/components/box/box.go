package box

import "nu/internal/tui/core"

// Box applies padding and background around child components.
type Box struct {
	core.Container
	opts Options
}

// New creates a box.
func New(opts Options) *Box {
	return &Box{opts: opts}
}
