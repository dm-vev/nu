package prompts

import (
	"context"
	"fmt"
)

// PromptManager manages prompt templates.
type PromptManager struct {
	store PromptStore
}

// NewPromptManager creates a prompt manager.
func NewPromptManager(store PromptStore) *PromptManager {
	return &PromptManager{
		store: store,
	}
}

// Get retrieves a template by ID and version
func (m *PromptManager) Get(ctx context.Context, id string, version string) (*PromptTemplate, error) {
	return m.store.Get(ctx, id, version)
}

// GetLatest retrieves the latest version of a template by ID
func (m *PromptManager) GetLatest(ctx context.Context, id string) (*PromptTemplate, error) {
	templates, err := m.store.List(ctx, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	// Find the latest version
	latest := templates[0]
	for _, tmpl := range templates[1:] {
		if tmpl.Version > latest.Version {
			latest = tmpl
		}
	}

	return latest, nil
}

// List returns all templates matching the given filter
func (m *PromptManager) List(ctx context.Context, filter map[string]interface{}) ([]*PromptTemplate, error) {
	return m.store.List(ctx, filter)
}

// Save stores a template
func (m *PromptManager) Save(ctx context.Context, tmpl *PromptTemplate) error {
	return m.store.Save(ctx, tmpl)
}

// Delete removes a template
func (m *PromptManager) Delete(ctx context.Context, id string, version string) error {
	return m.store.Delete(ctx, id, version)
}

// Render renders a template with the given data
func (m *PromptManager) Render(ctx context.Context, id string, version string, data map[string]interface{}) (string, error) {
	tmpl, err := m.Get(ctx, id, version)
	if err != nil {
		return "", err
	}

	return tmpl.Render(data)
}

// RenderLatest renders the latest version of a template with the given data
func (m *PromptManager) RenderLatest(ctx context.Context, id string, data map[string]interface{}) (string, error) {
	tmpl, err := m.GetLatest(ctx, id)
	if err != nil {
		return "", err
	}

	return tmpl.Render(data)
}
