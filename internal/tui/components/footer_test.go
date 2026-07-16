package components

import (
	"github.com/dm-vev/nu/internal/tui/ansi"
	"strings"
	"testing"
)

func TestFooterFooterFormatsPathAndStats(t *testing.T) {
	f := NewFooter(FooterOptions{
		CWD:      "/home/tikhon/Документы/nu",
		Home:     "/home/tikhon",
		Branch:   "main",
		Provider: "fireworks",
		Model:    "GLM 5.2 Fast",
		Used:     1280,
		Context:  128000,
	})

	lines := f.Render(60)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "~") || !strings.Contains(joined, "(main)") {
		t.Fatalf("footer = %q, want home-relative branch path", joined)
	}
	if !strings.Contains(joined, "1.0%/128k") || !strings.Contains(joined, "fireworks/GLM 5.2 Fast") {
		t.Fatalf("footer = %q, want context and model", joined)
	}
	for _, line := range lines {
		if ansi.VisibleWidth(line) != 60 {
			t.Fatalf("line width = %d, want 60 for %q", ansi.VisibleWidth(line), line)
		}
	}
}
