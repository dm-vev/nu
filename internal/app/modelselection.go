package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"nu/internal/app/auth"
	"nu/internal/app/cli"
	"nu/internal/model"
)

func loadModelRegistry(path string) ([]model.Model, model.Registry, error) {
	entries := model.Builtins()
	if strings.TrimSpace(path) != "" {
		custom, err := model.LoadCustom(path)
		if err != nil {
			return nil, model.Registry{}, err
		}
		entries = append(entries, custom...)
	}
	return entries, model.NewRegistry(entries), nil
}

func modelsPath(home string, explicit string) string {
	if strings.TrimSpace(explicit) != "" {
		return explicit
	}
	if strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".nu", "agent", "models.json")
}

func providerAuthState(ctx context.Context, store auth.Store, entries []model.Model) (map[string]bool, error) {
	state := map[string]bool{}
	seen := map[string]bool{}
	for _, entry := range entries {
		if seen[entry.Provider] {
			continue
		}
		seen[entry.Provider] = true
		// Auth resolution may run configured commands, so do it once per provider.
		_, ok, err := store.ResolveAPIKey(ctx, entry.Provider)
		if err != nil {
			return nil, err
		}
		if ok {
			state[entry.Provider] = true
		}
	}
	return state, nil
}

func markConfiguredProviders(state map[string]bool, entries []model.Model) {
	for _, entry := range entries {
		state[entry.Provider] = true
	}
}

func selectModel(
	registry model.Registry,
	authState map[string]bool,
	req cli.Request,
	settings providerSettingsFile,
) (model.Model, error) {
	providerID := strings.TrimSpace(req.Provider)
	modelID := strings.TrimSpace(req.Model)
	if modelID != "" {
		if providerID != "" {
			return selectProviderModel(registry, authState, providerID, modelID)
		}
		return registry.Resolve(modelID, authState)
	}

	available := registry.Available(authState)
	if providerID != "" {
		if selected, ok := configuredDefaultForProvider(registry, authState, providerID, settings); ok {
			return selected, nil
		}
		if selected, ok := defaultModelForProvider(registry, authState, providerID); ok {
			return selected, nil
		}
		for _, entry := range available {
			if entry.Provider == providerID {
				return entry, nil
			}
		}
		return model.Model{}, fmt.Errorf("resolve provider %q: no available models", providerID)
	}
	if len(available) == 0 {
		return model.Model{}, fmt.Errorf("resolve model: no available models")
	}
	if selected, ok := configuredDefault(registry, authState, settings); ok {
		return selected, nil
	}
	// The global default should be stable instead of depending on registry sort order.
	if selected, err := registry.Resolve("openai-default", authState); err == nil {
		return selected, nil
	}
	return available[0], nil
}

func configuredDefault(registry model.Registry, authState map[string]bool, settings providerSettingsFile) (model.Model, bool) {
	providerID := strings.TrimSpace(settings.DefaultProvider)
	modelID := strings.TrimSpace(settings.DefaultModel)
	if providerID != "" && modelID != "" {
		if selected, err := selectProviderModel(registry, authState, providerID, modelID); err == nil {
			return selected, true
		}
	}
	if providerID != "" {
		return configuredDefaultForProvider(registry, authState, providerID, settings)
	}
	if modelID != "" {
		if selected, err := registry.Resolve(modelID, authState); err == nil {
			return selected, true
		}
	}
	return model.Model{}, false
}

func configuredDefaultForProvider(
	registry model.Registry,
	authState map[string]bool,
	providerID string,
	settings providerSettingsFile,
) (model.Model, bool) {
	setting, ok := settings.Providers[providerID]
	if !ok || strings.TrimSpace(setting.DefaultModel) == "" {
		return model.Model{}, false
	}
	selected, err := selectProviderModel(registry, authState, providerID, setting.DefaultModel)
	return selected, err == nil
}

func defaultModelForProvider(registry model.Registry, authState map[string]bool, providerID string) (model.Model, bool) {
	var selector string
	switch providerID {
	case "openai":
		selector = "openai-default"
	default:
		return model.Model{}, false
	}
	selected, err := registry.Resolve(selector, authState)
	if err != nil || selected.Provider != providerID {
		return model.Model{}, false
	}
	return selected, true
}

func selectProviderModel(registry model.Registry, authState map[string]bool, providerID string, modelID string) (model.Model, error) {
	if strings.Contains(modelID, "/") {
		selected, err := registry.Resolve(modelID, authState)
		if err != nil {
			return model.Model{}, err
		}
		if selected.Provider != providerID {
			return model.Model{}, fmt.Errorf("resolve model %q: provider is %s, want %s", modelID, selected.Provider, providerID)
		}
		return selected, nil
	}
	return registry.Resolve(providerID+"/"+modelID, authState)
}
