package modelmenu

// Action is the semantic result of one selector input event.
type Action int

const (
	ActionNone Action = iota
	ActionChanged
	ActionSelect
	ActionCancel
)

// HandleInput updates selector state from a raw terminal event.
func (m *Menu) HandleInput(data string) Action {
	if !m.visible {
		return ActionNone
	}
	switch data {
	case "\x1b[A":
		m.move(-1)
		return ActionChanged
	case "\x1b[B":
		m.move(1)
		return ActionChanged
	case "\r", "\n":
		return ActionSelect
	case "\x1b", "\x03":
		return ActionCancel
	case "\x7f", "\b":
		m.backspace()
		return ActionChanged
	}
	if !isPrintable(data) {
		return ActionNone
	}
	m.query += data
	m.refresh()
	return ActionChanged
}

func (m *Menu) move(delta int) {
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

func (m *Menu) backspace() {
	runes := []rune(m.query)
	if len(runes) == 0 {
		return
	}
	m.query = string(runes[:len(runes)-1])
	m.refresh()
}

func isPrintable(data string) bool {
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
