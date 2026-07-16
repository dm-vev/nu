package components

// Status renders transient agent state.
type Status struct {
	text   string
	style  func(string) string
	frame  int
	alert  bool
	frames []string
}

// NewStatus creates an empty status line.
func NewStatus(style func(string) string, frameSet ...string) *Status {
	if style == nil {
		style = statusIdentity
	}
	frames := statusDefaultFrames
	if len(frameSet) > 0 {
		frames = append([]string(nil), frameSet...)
	}
	return &Status{style: style, frames: frames}
}

// SetText replaces status text.
func (s *Status) SetText(value string) {
	s.text = value
	s.frame = 0
	s.alert = false
}

// SetAlertText replaces status text and renders it with the alert gradient.
func (s *Status) SetAlertText(value string) {
	s.text = value
	s.frame = 0
	s.alert = true
}

// Text returns the raw status text.
func (s *Status) Text() string {
	return s.text
}

// Step advances the lightweight working animation.
func (s *Status) Step() {
	s.frame = (s.frame + 1) % len(s.frames)
}

// Invalidate exists for the component interface.
func (s *Status) Invalidate() {}

var statusDefaultFrames = []string{"·", "✢", "✳", "✶", "✻", "✽"}

func statusIdentity(value string) string {
	return value
}
