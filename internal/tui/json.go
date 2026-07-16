package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeBoolMap(path string, key string, value bool) error {
	values := map[string]bool{}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &values)
	}
	values[key] = value
	return writeJSONFile(path, values)
}

func writeAuthProvider(path string, providerID string, credential map[string]string) error {
	file := map[string]map[string]map[string]string{"providers": {}}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &file)
	}
	if file["providers"] == nil {
		file["providers"] = map[string]map[string]string{}
	}
	file["providers"][providerID] = credential
	return writeJSONFile(path, file)
}

func removeAuthProvider(path string, providerID string) error {
	file := map[string]map[string]map[string]string{"providers": {}}
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &file)
	}
	if file["providers"] == nil {
		file["providers"] = map[string]map[string]string{}
	}
	delete(file["providers"], providerID)
	return writeJSONFile(path, file)
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir %s: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode json %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write json %s: %w", path, err)
	}
	return nil
}
