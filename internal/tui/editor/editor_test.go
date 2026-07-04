package editor

import (
	"strings"
	"testing"

	"nu/internal/tui/ansi"
)

func TestEditorWrapsInput(t *testing.T) {
	e := New()
	e.HandleInput("alpha beta gamma delta")
	lines := e.Render(12)
	plain := strings.Join(lines, "\n")
	if !strings.Contains(plain, "alpha beta") || !strings.Contains(plain, "gamma delta") {
		t.Fatalf("lines = %#v", lines)
	}
	for _, line := range lines {
		if ansi.VisibleWidth(line) > 12 {
			t.Fatalf("line width = %d: %q", ansi.VisibleWidth(line), line)
		}
	}
}

func TestEditorTrailingSpaceKeepsCursorOnSameVisualLine(t *testing.T) {
	e := New()
	e.HandleInput("hello")
	before := len(e.Render(20))

	e.HandleInput(" ")
	after := len(e.Render(20))

	if after != before {
		t.Fatalf("rendered lines after trailing space = %d, want %d", after, before)
	}
}

func TestEditorCanUseASCIIBorder(t *testing.T) {
	e := New()
	e.SetBorderRune('-')

	lines := e.Render(8)

	if lines[0] != "--------" || lines[len(lines)-1] != "--------" {
		t.Fatalf("lines = %#v, want ASCII border", lines)
	}
}

func TestEditorSubmit(t *testing.T) {
	e := New()
	var submitted string
	e.SetSubmitHandler(func(text string) { submitted = text })
	e.HandleInput("hello")
	e.HandleInput("\n")
	if submitted != "hello" {
		t.Fatalf("submitted = %q", submitted)
	}
	if e.Text() != "" {
		t.Fatalf("Text after submit = %q", e.Text())
	}
}

func TestEditorHandlesUnicodeCursorAndPaste(t *testing.T) {
	e := New()

	e.HandleInput("п")
	e.HandleInput("р")
	e.HandleInput("\x1b[D")
	e.HandleInput("и")
	if e.Text() != "пир" {
		t.Fatalf("text = %q, want unicode insertion at cursor", e.Text())
	}

	e.HandleInput(bracketedPasteStart + "вет\nмир" + bracketedPasteEnd)
	if e.Text() != "пивет\nмирр" {
		t.Fatalf("text = %q, want pasted multiline text", e.Text())
	}
}

func TestEditorForwardDeleteUsesRuneCursor(t *testing.T) {
	e := New()
	e.HandleInput("аб")
	e.HandleInput("\x1b[D")
	e.HandleInput("\x1b[3~")

	if e.Text() != "а" {
		t.Fatalf("text = %q, want rune forward delete", e.Text())
	}
}
