package commandmenu

import "nu/internal/slash"

// Menu renders slash command suggestions.
type Menu struct {
	commands []slash.Command
	matches  []slash.Command
	prefix   string
	selected int
	opts     Options
}

// New creates an idle command menu.
func New(commands []slash.Command, opts Options) *Menu {
	return &Menu{commands: append([]slash.Command(nil), commands...), opts: normalizeOptions(opts)}
}

// SetText updates visible suggestions from editor text.
func (m *Menu) SetText(text string) {
	prefix, ok := menuPrefix(text)
	if !ok {
		m.prefix = ""
		m.matches = nil
		m.selected = 0
		return
	}
	if prefix != m.prefix {
		m.selected = 0
	}
	m.prefix = prefix
	m.matches = slash.Filter(prefix, m.opts.MaxItems)
	if m.selected >= len(m.matches) {
		m.selected = max(0, len(m.matches)-1)
	}
}

// Completion returns the highlighted command completion.
func (m *Menu) Completion() (string, bool) {
	command, ok := m.Selected()
	if !ok {
		return "", false
	}
	return "/" + command.Name + " ", true
}

// Selected returns the highlighted command.
func (m *Menu) Selected() (slash.Command, bool) {
	if m.selected < 0 || m.selected >= len(m.matches) {
		return slash.Command{}, false
	}
	return m.matches[m.selected], true
}

// Move changes the highlighted command with wraparound.
func (m *Menu) Move(delta int) bool {
	if len(m.matches) == 0 {
		return false
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = len(m.matches) - 1
	}
	if m.selected >= len(m.matches) {
		m.selected = 0
	}
	return true
}

// Visible reports whether the command selector is active.
func (m *Menu) Visible() bool {
	return len(m.matches) > 0
}

// Invalidate exists for the component interface.
func (m *Menu) Invalidate() {}
