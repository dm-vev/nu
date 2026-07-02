package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNUF031ModelPatternSelectsProviderAndModel(t *testing.T) {
	registry := NewRegistry(Builtins())

	got, err := registry.Resolve("openai/gpt-5*", map[string]bool{"openai": true})
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if got.Provider != "openai" || got.API != "responses" {
		t.Fatalf("model=%#v, want OpenAI Responses model", got)
	}
}

func TestNUF031UnavailableModelsHiddenWithoutAuth(t *testing.T) {
	registry := NewRegistry(Builtins())

	models := registry.Available(nil)
	for _, got := range models {
		if got.Provider == "openai" || got.Provider == "anthropic" {
			t.Fatalf("available model %#v should be hidden without auth", got)
		}
	}
}

func TestNUF031CustomModelsOverrideBuiltins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.json")
	raw := `{"models":[{"id":"gpt-5.5","provider":"openai","api":"responses","context_window":123}]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	custom, err := LoadCustom(path)
	if err != nil {
		t.Fatalf("LoadCustom error = %v", err)
	}
	registry := NewRegistry(append(Builtins(), custom...))
	got, err := registry.Resolve("openai/gpt-5.5", map[string]bool{"openai": true})
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if got.ContextWindow != 123 {
		t.Fatalf("ContextWindow=%d, want custom override", got.ContextWindow)
	}
}

func TestCustomModelsCanDisableEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "models.json")
	raw := `{"models":[{"id":"gpt-5.5","provider":"openai","api":"responses","enabled":false}]}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	custom, err := LoadCustom(path)
	if err != nil {
		t.Fatalf("LoadCustom error = %v", err)
	}
	registry := NewRegistry(append(Builtins(), custom...))
	if _, err := registry.Resolve("openai/gpt-5.5", map[string]bool{"openai": true}); err == nil {
		t.Fatalf("Resolve error = nil, want disabled model hidden")
	}
}

func TestNUF032ThinkingLevelMapping(t *testing.T) {
	got, err := ThinkingFor("openai", "responses", ThinkingHigh)
	if err != nil {
		t.Fatalf("ThinkingFor error = %v", err)
	}
	reasoning, ok := got["reasoning"].(map[string]any)
	if !ok || reasoning["effort"] != "high" {
		t.Fatalf("thinking=%#v, want OpenAI reasoning effort high", got)
	}
}

func TestNUF032UnsupportedThinkingLevelFallsBackOrErrors(t *testing.T) {
	_, err := ThinkingFor("openai", "chat", ThinkingHigh)
	if err == nil {
		t.Fatalf("ThinkingFor error = nil, want unsupported error")
	}
}
