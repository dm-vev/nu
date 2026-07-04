package thinking

import (
	"strings"
	"testing"

	"nu/internal/tui/ansi"
)

func TestThinkingRendersMarkdownWithThinkingStyle(t *testing.T) {
	component := New("model **reasoning**", Options{
		PaddingX: 1,
		TextStyle: func(value string) string {
			return ansi.Italic + ansi.Dim + value + ansi.BoldOff + ansi.ItalicOff + ansi.DefaultFG
		},
		StrongStyle: func(value string) string {
			return ansi.Italic + ansi.Dim + ansi.Bold + value + ansi.BoldOff + ansi.ItalicOff + ansi.DefaultFG
		},
	})

	joined := strings.Join(component.Render(60), "\n")
	if !strings.Contains(joined, ansi.Italic) || !strings.Contains(joined, ansi.Dim) {
		t.Fatalf("rendered thinking = %q, want italic dim style", joined)
	}
	if !strings.Contains(ansi.Strip(joined), "model reasoning") {
		t.Fatalf("rendered thinking = %q, want visible text", joined)
	}
}
