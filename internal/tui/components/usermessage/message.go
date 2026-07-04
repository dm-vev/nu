package usermessage

// Message renders one user turn.
type Message struct {
	text string
	opts Options
}

// New creates a user message component.
func New(value string, opts Options) *Message {
	return &Message{text: value, opts: normalizeOptions(opts)}
}

// SetText replaces message text.
func (m *Message) SetText(value string) {
	m.text = value
}

// Text returns raw message text.
func (m *Message) Text() string {
	return m.text
}

// Invalidate exists for the component interface.
func (m *Message) Invalidate() {}
