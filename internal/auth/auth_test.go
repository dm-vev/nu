package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNUF020AuthFileBeatsEnvironment(t *testing.T) {
	path := writeAuth(t, `{"providers":{"openai":{"api_key":"file-key"}}}`)

	store, err := Load(path, []string{"OPENAI_API_KEY=env-key"})
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}
	key, ok, err := store.ResolveAPIKey(context.Background(), "openai")
	if err != nil {
		t.Fatalf("ResolveAPIKey error = %v", err)
	}
	if !ok || key != "file-key" {
		t.Fatalf("key=%q ok=%v, want file-key true", key, ok)
	}
}

func TestNUF020EnvInterpolation(t *testing.T) {
	path := writeAuth(t, `{"providers":{"openai":{"api_key":"${OPENAI_API_KEY}-suffix"}}}`)

	store, err := Load(path, []string{"OPENAI_API_KEY=env-key"})
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}
	key, ok, err := store.ResolveAPIKey(context.Background(), "openai")
	if err != nil {
		t.Fatalf("ResolveAPIKey error = %v", err)
	}
	if !ok || key != "env-key-suffix" {
		t.Fatalf("key=%q ok=%v, want env-key-suffix true", key, ok)
	}
}

func TestNUF020CommandInterpolation(t *testing.T) {
	path := writeAuth(t, `{"providers":{"openai":{"api_key_command":"printf command-key"}}}`)

	store, err := Load(path, nil)
	if err != nil {
		t.Fatalf("Load error = %v", err)
	}
	key, ok, err := store.ResolveAPIKey(context.Background(), "openai")
	if err != nil {
		t.Fatalf("ResolveAPIKey error = %v", err)
	}
	if !ok || key != "command-key" {
		t.Fatalf("key=%q ok=%v, want command-key true", key, ok)
	}
}

func writeAuth(t *testing.T, raw string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	return path
}
