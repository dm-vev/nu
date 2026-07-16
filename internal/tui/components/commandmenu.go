package components

// CommandMenu renders slash command suggestions.
type CommandMenu struct {
	commands []SlashCommand
	matches  []SlashCommand
	prefix   string
	selected int
	opts     CommandMenuOptions
}

// NewCommandMenu creates an idle command menu.
func NewCommandMenu(commands []SlashCommand, opts CommandMenuOptions) *CommandMenu {
	return &CommandMenu{commands: append([]SlashCommand(nil), commands...), opts: commandMenuNormalizeOptions(opts)}
}

// SetText updates visible suggestions from editor text.
func (m *CommandMenu) SetText(text string) {
	prefix, ok := commandMenuMenuPrefix(text)
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
	m.matches = SlashFilter(prefix, m.opts.MaxItems)
	if m.selected >= len(m.matches) {
		m.selected = max(0, len(m.matches)-1)
	}
}

// Completion returns the highlighted command completion.
func (m *CommandMenu) Completion() (string, bool) {
	command, ok := m.Selected()
	if !ok {
		return "", false
	}
	return "/" + command.Name + " ", true
}

// Selected returns the highlighted command.
func (m *CommandMenu) Selected() (SlashCommand, bool) {
	if m.selected < 0 || m.selected >= len(m.matches) {
		return SlashCommand{}, false
	}
	return m.matches[m.selected], true
}

// Move changes the highlighted command with wraparound.
func (m *CommandMenu) Move(delta int) bool {
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
func (m *CommandMenu) Visible() bool {
	return len(m.matches) > 0
}

// Invalidate exists for the component interface.
func (m *CommandMenu) Invalidate() {}
