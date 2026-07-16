package components

import (
	"nu/internal/tui/ansi"
	"strings"
	"testing"
)

func TestCommandMenuMenuRendersFilteredCommands(t *testing.T) {
	menu := NewCommandMenu(SlashBuiltins(), CommandMenuOptions{})
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

func TestCommandMenuMenuMovesSelectionAndCompletesSelectedCommand(t *testing.T) {
	menu := NewCommandMenu(SlashBuiltins(), CommandMenuOptions{})
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

func TestCommandMenuMenuHidesOutsideSlashPrefix(t *testing.T) {
	menu := NewCommandMenu(SlashBuiltins(), CommandMenuOptions{})
	menu.SetText("hello /mo")
	if lines := menu.Render(72); len(lines) != 0 {
		t.Fatalf("lines = %#v, want hidden menu", lines)
	}
}
