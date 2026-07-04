package commandmenu

import (
	"strings"
	"testing"

	"nu/internal/slash"
	"nu/internal/tui/ansi"
)

func TestMenuRendersFilteredCommands(t *testing.T) {
	menu := New(slash.Builtins(), Options{})
	menu.SetText("/mo")

	lines := menu.Render(72)
	plain := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(plain, "/model") || !strings.Contains(plain, "/scoped-models") {
		t.Fatalf("menu = %q, want model commands", plain)
	}
	if completion, ok := menu.Completion(); !ok || completion != "/model " {
		t.Fatalf("Completion = %q, %v", completion, ok)
	}
}

func TestMenuMovesSelectionAndCompletesSelectedCommand(t *testing.T) {
	menu := New(slash.Builtins(), Options{})
	menu.SetText("/")

	if !menu.Move(1) {
		t.Fatalf("Move returned false, want true")
	}
	completion, ok := menu.Completion()
	if !ok || completion != "/model " {
		t.Fatalf("Completion = %q, %v, want /model", completion, ok)
	}

	lines := menu.Render(72)
	plain := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.Contains(plain, "> /model") {
		t.Fatalf("menu = %q, want selected model row", plain)
	}
}

func TestMenuHidesOutsideSlashPrefix(t *testing.T) {
	menu := New(slash.Builtins(), Options{})
	menu.SetText("hello /mo")
	if lines := menu.Render(72); len(lines) != 0 {
		t.Fatalf("lines = %#v, want hidden menu", lines)
	}
}
