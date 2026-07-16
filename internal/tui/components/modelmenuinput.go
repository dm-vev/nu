package components

// ModelMenuAction is the semantic result of one selector input event.
type ModelMenuAction int

const (
	ModelMenuActionNone ModelMenuAction = iota
	ModelMenuActionChanged
	ModelMenuActionSelect
	ModelMenuActionCancel
)

// HandleInput updates selector state from a raw terminal event.
func (m *ModelMenu) HandleInput(data string) ModelMenuAction {
	if !m.visible {
		return ModelMenuActionNone
	}
	switch data {
	case "\x1b[A":
		m.move(-1)
		return ModelMenuActionChanged
	case "\x1b[B":
		m.move(1)
		return ModelMenuActionChanged
	case "\r", "\n":
		return ModelMenuActionSelect
	case "\x1b", "\x03":
		return ModelMenuActionCancel
	case "\x7f", "\b":
		m.backspace()
		return ModelMenuActionChanged
	}
	if !modelMenuIsPrintable(data) {
		return ModelMenuActionNone
	}
	m.query += data
	m.refresh()
	return ModelMenuActionChanged
}

func (m *ModelMenu) move(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = len(m.filtered) - 1
		return
	}
	if m.selected >= len(m.filtered) {
		m.selected = 0
	}
}

func (m *ModelMenu) backspace() {
	runes := []rune(m.query)
	if len(runes) == 0 {
		return
	}
	m.query = string(runes[:len(runes)-1])
	m.refresh()
}

func modelMenuIsPrintable(data string) bool {
	if data == "" {
		return false
	}
	for _, char := range data {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}
