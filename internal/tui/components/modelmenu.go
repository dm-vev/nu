package components

import "github.com/dm-vev/nu/internal/model"

// ModelMenu renders the Pi-style /model selector.
type ModelMenu struct {
	models          []model.Model
	filtered        []model.Model
	query           string
	currentProvider string
	currentID       string
	selected        int
	visible         bool
	opts            ModelMenuOptions
}

// NewModelMenu creates a hidden model selector.
func NewModelMenu(models []model.Model, opts ModelMenuOptions) *ModelMenu {
	menu := &ModelMenu{opts: modelMenuNormalizeOptions(opts)}
	menu.SetModels(models)
	return menu
}

// SetModels replaces the available model list.
func (m *ModelMenu) SetModels(models []model.Model) {
	m.models = append([]model.Model(nil), models...)
	if m.visible {
		m.refresh()
	}
}

// Open shows the selector with an optional search query.
func (m *ModelMenu) Open(query string, currentProvider string, currentID string) {
	m.query = query
	m.currentProvider = currentProvider
	m.currentID = currentID
	m.visible = true
	m.refresh()
}

// Close hides the selector and clears transient search state.
func (m *ModelMenu) Close() {
	m.visible = false
	m.query = ""
	m.filtered = nil
	m.selected = 0
}

// Visible reports whether the selector should receive input and render.
func (m *ModelMenu) Visible() bool {
	return m.visible
}

// Query returns the active search text.
func (m *ModelMenu) Query() string {
	return m.query
}

// Selected returns the currently highlighted model.
func (m *ModelMenu) Selected() (model.Model, bool) {
	if m.selected < 0 || m.selected >= len(m.filtered) {
		return model.Model{}, false
	}
	return m.filtered[m.selected], true
}

// Invalidate exists for the component interface.
func (m *ModelMenu) Invalidate() {}

func (m *ModelMenu) isCurrent(entry model.Model) bool {
	return entry.Provider == m.currentProvider && entry.ID == m.currentID
}
