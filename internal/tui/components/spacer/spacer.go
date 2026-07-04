package spacer

// Spacer renders blank lines.
type Spacer struct {
	lines int
}

// New creates a spacer.
func New(lines int) *Spacer {
	return &Spacer{lines: lines}
}

// SetLines updates spacer height.
func (s *Spacer) SetLines(lines int) {
	s.lines = lines
}
