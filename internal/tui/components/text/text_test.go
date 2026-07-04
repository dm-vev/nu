package text

import (
	"strings"
	"testing"

	"nu/internal/tui/ansi"
)

func TestTextRenderWrapsAndPads(t *testing.T) {
	component := New("alpha beta gamma delta", Options{PaddingX: 1})
	lines := component.Render(14)

	plain := strings.Join(lines, "\n")
	if !strings.Contains(plain, " alpha beta ") || !strings.Contains(plain, " gamma delta ") {
		t.Fatalf("lines = %#v", lines)
	}
	for _, line := range lines {
		if ansi.VisibleWidth(line) != 14 {
			t.Fatalf("line width = %d, want 14: %q", ansi.VisibleWidth(line), line)
		}
	}
}
