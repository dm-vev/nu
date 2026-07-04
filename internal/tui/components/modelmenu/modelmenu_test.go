package modelmenu

import (
	"strings"
	"testing"

	"nu/internal/model"
)

func TestModelMenuRendersCurrentModelFirst(t *testing.T) {
	menu := New([]model.Model{
		{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
		{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
	}, Options{})

	menu.Open("", "fireworks", "glm-fast")

	lines := menu.Render(80)
	plain := strings.Join(lines, "\n")
	if !strings.Contains(plain, "> glm-fast [fireworks] *") {
		t.Fatalf("render = %q, want current model first and marked", plain)
	}
}

func TestModelMenuFiltersByDisplayNameAndSelects(t *testing.T) {
	menu := New([]model.Model{
		{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
		{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
	}, Options{})

	menu.Open("fast", "openai", "gpt-test")

	selected, ok := menu.Selected()
	if !ok {
		t.Fatalf("Selected ok = false, want true")
	}
	if selected.Provider != "fireworks" || selected.ID != "glm-fast" {
		t.Fatalf("selected = %s/%s, want fireworks/glm-fast", selected.Provider, selected.ID)
	}

	if action := menu.HandleInput("\r"); action != ActionSelect {
		t.Fatalf("HandleInput action = %d, want ActionSelect", action)
	}
}
