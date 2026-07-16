package components

// UserMessage renders one user turn.
type UserMessage struct {
	text string
	opts UserMessageOptions
}

// NewUserMessage creates a user message component.
func NewUserMessage(value string, opts UserMessageOptions) *UserMessage {
	return &UserMessage{text: value, opts: userMessageNormalizeOptions(opts)}
}

// SetText replaces message text.
func (m *UserMessage) SetText(value string) {
	m.text = value
}

// Text returns raw message text.
func (m *UserMessage) Text() string {
	return m.text
}

// Invalidate exists for the component interface.
func (m *UserMessage) Invalidate() {}
