package components

import "nu/internal/tui/message"

import (
	"strings"
)

// AssistantMessage renders one assistant turn.
type AssistantMessage struct {
	parts []message.Part
	opts  AssistantMessageOptions
}

// NewAssistantMessage creates an assistant message component.
func NewAssistantMessage(value string, opts AssistantMessageOptions) *AssistantMessage {
	return &AssistantMessage{
		parts: []message.Part{{Kind: message.PartText, Text: value}},
		opts:  assistantMessageNormalizeOptions(opts),
	}
}

// NewAssistantMessageFromMessage creates an assistant message component from structured content.
func NewAssistantMessageFromMessage(value message.Message, opts AssistantMessageOptions) *AssistantMessage {
	return &AssistantMessage{parts: append([]message.Part(nil), value.Parts...), opts: assistantMessageNormalizeOptions(opts)}
}

// SetText replaces message text.
func (m *AssistantMessage) SetText(value string) {
	m.parts = []message.Part{{Kind: message.PartText, Text: value}}
}

// Text returns raw message text.
func (m *AssistantMessage) Text() string {
	values := make([]string, 0, len(m.parts))
	for _, part := range m.parts {
		if part.Kind == message.PartText {
			values = append(values, part.Text)
		}
	}
	return strings.Join(values, "")
}

// Invalidate exists for the component interface.
func (m *AssistantMessage) Invalidate() {}
