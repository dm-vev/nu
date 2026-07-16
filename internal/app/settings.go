package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nu/internal/model"
)

type providerSetting struct {
	BaseURL      string `json:"base_url,omitempty"`
	API          string `json:"api,omitempty"`
	AuthProvider string `json:"auth_provider,omitempty"`
	ModelsFile   string `json:"models_file,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
}

type providerSettingsFile struct {
	DefaultProvider string                     `json:"default_provider,omitempty"`
	DefaultModel    string                     `json:"default_model,omitempty"`
	Providers       map[string]providerSetting `json:"providers"`
}

func loadProviderSettings(home string) (providerSettingsFile, error) {
	settings := providerSettingsFile{Providers: map[string]providerSetting{}}
	if strings.TrimSpace(home) == "" {
		return settings, nil
	}
	path := filepath.Join(home, ".nu", "agent", "settings.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return providerSettingsFile{}, fmt.Errorf("read provider settings: %w", err)
	}
	var file providerSettingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return providerSettingsFile{}, fmt.Errorf("decode provider settings: %w", err)
	}
	if file.Providers == nil {
		file.Providers = map[string]providerSetting{}
	}
	return file, nil
}

func saveSelectedModel(home string, selected model.Model) error {
	if strings.TrimSpace(home) == "" {
		return nil
	}
	settings, err := loadProviderSettings(home)
	if err != nil {
		return err
	}
	settings.DefaultProvider = selected.Provider
	settings.DefaultModel = selected.ID
	if settings.Providers == nil {
		settings.Providers = map[string]providerSetting{}
	}
	providerSettings := settings.Providers[selected.Provider]
	providerSettings.DefaultModel = selected.ID
	if providerSettings.API == "" {
		providerSettings.API = selected.API
	}
	settings.Providers[selected.Provider] = providerSettings
	return writeProviderSettings(home, settings)
}

func writeProviderSettings(home string, settings providerSettingsFile) error {
	dir := filepath.Join(home, ".nu", "agent")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create provider settings dir: %w", err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("encode provider settings: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(dir, "settings.json")
	tmp, err := os.CreateTemp(dir, "settings.json.*")
	if err != nil {
		return fmt.Errorf("create provider settings temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write provider settings temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close provider settings temp: %w", err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod provider settings temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("replace provider settings: %w", err)
	}
	return nil
}
