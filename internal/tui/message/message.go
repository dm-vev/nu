package message

// Message is one chat turn split into renderable parts.
type Message struct {
	Role  string
	Parts []Part
}

// NewUser returns a user message with Markdown-capable text.
func NewUser(text string) Message {
	return Message{
		Role:  RoleUser,
		Parts: []Part{{Kind: PartText, Text: text}},
	}
}

// NewAssistant returns an empty assistant turn.
func NewAssistant() Message {
	return Message{Role: RoleAssistant}
}

// NewAssistantText returns an assistant turn with visible text.
func NewAssistantText(text string) Message {
	msg := NewAssistant()
	msg.AppendText(text)
	return msg
}

// Clone returns a copy whose parts can be changed independently.
func (m Message) Clone() Message {
	parts := append([]Part(nil), m.Parts...)
	return Message{Role: m.Role, Parts: parts}
}
