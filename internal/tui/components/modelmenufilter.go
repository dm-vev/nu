package components

import (
	"sort"
	"strings"

	"github.com/dm-vev/nu/internal/model"
)

func (m *ModelMenu) refresh() {
	m.filtered = m.filtered[:0]
	for _, entry := range m.models {
		if modelMenuMatchesQuery(entry, m.query) {
			m.filtered = append(m.filtered, entry)
		}
	}

	// Keep the active model at the top like Pi, then preserve deterministic provider/id ordering.
	sort.SliceStable(m.filtered, func(i, j int) bool {
		left := m.filtered[i]
		right := m.filtered[j]
		leftCurrent := m.isCurrent(left)
		rightCurrent := m.isCurrent(right)
		if leftCurrent != rightCurrent {
			return leftCurrent
		}
		if left.Provider != right.Provider {
			return left.Provider < right.Provider
		}
		return left.ID < right.ID
	})

	if len(m.filtered) == 0 {
		m.selected = 0
		return
	}
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func modelMenuMatchesQuery(entry model.Model, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return true
	}
	search := strings.ToLower(modelMenuModelSearchText(entry))
	for _, token := range strings.Fields(query) {
		if !strings.Contains(search, token) {
			return false
		}
	}
	return true
}

func modelMenuModelSearchText(entry model.Model) string {
	parts := []string{entry.Provider, entry.ID, entry.Provider + "/" + entry.ID, entry.DisplayName}
	parts = append(parts, entry.Aliases...)
	return strings.Join(parts, " ")
}

func modelMenuModelDisplayName(entry model.Model) string {
	if strings.TrimSpace(entry.DisplayName) != "" {
		return entry.DisplayName
	}
	return entry.ID
}
