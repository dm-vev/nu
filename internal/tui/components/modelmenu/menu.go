package modelmenu

import "nu/internal/model"

// Menu renders the Pi-style /model selector.
type Menu struct {
	models          []model.Model
	filtered        []model.Model
	query           string
	currentProvider string
	currentID       string
	selected        int
	visible         bool
	opts            Options
}

// New creates a hidden model selector.
func New(models []model.Model, opts Options) *Menu {
	menu := &Menu{opts: normalizeOptions(opts)}
	menu.SetModels(models)
	return menu
}

// SetModels replaces the available model list.
func (m *Menu) SetModels(models []model.Model) {
	m.models = append([]model.Model(nil), models...)
	if m.visible {
		m.refresh()
	}
}

// Open shows the selector with an optional search query.
func (m *Menu) Open(query string, currentProvider string, currentID string) {
	m.query = query
	m.currentProvider = currentProvider
	m.currentID = currentID
	m.visible = true
	m.refresh()
}

// Close hides the selector and clears transient search state.
func (m *Menu) Close() {
	m.visible = false
	m.query = ""
	m.filtered = nil
	m.selected = 0
}

// Visible reports whether the selector should receive input and render.
func (m *Menu) Visible() bool {
	return m.visible
}

// Query returns the active search text.
func (m *Menu) Query() string {
	return m.query
}

// Selected returns the currently highlighted model.
func (m *Menu) Selected() (model.Model, bool) {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return model.Model{}, false
	}
	return m.filtered[m.selected], true
}

// Invalidate exists for the component interface.
func (m *Menu) Invalidate() {}

func (m *Menu) isCurrent(entry model.Model) bool {
	return entry.Provider == m.currentProvider && entry.ID == m.currentID
}
