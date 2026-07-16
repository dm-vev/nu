package components

// Spacer renders blank lines.
type Spacer struct {
	lines int
}

// NewSpacer creates a spacer.
func NewSpacer(lines int) *Spacer {
	return &Spacer{lines: lines}
}

// SetLines updates spacer height.
func (s *Spacer) SetLines(lines int) {
	s.lines = lines
}
