package assistantmessage

import (
	"strings"

	tuimessage "nu/internal/tui/message"
)

// Message renders one assistant turn.
type Message struct {
	parts []tuimessage.Part
	opts  Options
}

// New creates an assistant message component.
func New(value string, opts Options) *Message {
	return &Message{
		parts: []tuimessage.Part{{Kind: tuimessage.PartText, Text: value}},
		opts:  normalizeOptions(opts),
	}
}

// NewMessage creates an assistant message component from structured content.
func NewMessage(value tuimessage.Message, opts Options) *Message {
	return &Message{parts: append([]tuimessage.Part(nil), value.Parts...), opts: normalizeOptions(opts)}
}

// SetText replaces message text.
func (m *Message) SetText(value string) {
	m.parts = []tuimessage.Part{{Kind: tuimessage.PartText, Text: value}}
}

// Text returns raw message text.
func (m *Message) Text() string {
	values := make([]string, 0, len(m.parts))
	for _, part := range m.parts {
		if part.Kind == tuimessage.PartText {
			values = append(values, part.Text)
		}
	}
	return strings.Join(values, "")
}

// Invalidate exists for the component interface.
func (m *Message) Invalidate() {}
