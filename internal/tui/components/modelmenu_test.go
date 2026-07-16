package components

import (
	"strings"
	"testing"

	"github.com/dm-vev/nu/internal/model"
)

func TestModelMenuModelMenuRendersCurrentModelFirst(t *testing.T) {
	menu := NewModelMenu([]model.Model{
		{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
		{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
	}, ModelMenuOptions{})

	menu.Open("", "fireworks", "glm-fast")

	lines := menu.Render(80)
	plain := strings.Join(lines, "\n")
	if !strings.Contains(plain, "> glm-fast [fireworks] *") {
		t.Fatalf("render = %q, want current model first and marked", plain)
	}
}

func TestModelMenuModelMenuFiltersByDisplayNameAndSelects(t *testing.T) {
	menu := NewModelMenu([]model.Model{
		{ID: "gpt-test", Provider: "openai", API: "chat", DisplayName: "GPT Test", Enabled: true},
		{ID: "glm-fast", Provider: "fireworks", API: "chat", DisplayName: "GLM Fast", Enabled: true},
	}, ModelMenuOptions{})

	menu.Open("fast", "openai", "gpt-test")

	selected, ok := menu.Selected()
	if !ok {
		t.Fatalf("Selected ok = false, want true")
	}
	if selected.Provider != "fireworks" || selected.ID != "glm-fast" {
		t.Fatalf("selected = %s/%s, want fireworks/glm-fast", selected.Provider, selected.ID)
	}

	if action := menu.HandleInput("\r"); action != ModelMenuActionSelect {
		t.Fatalf("HandleInput action = %d, want ActionSelect", action)
	}
}
