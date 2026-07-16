package components

import (
	"strings"
	"testing"
)

func TestHeaderHeaderRendersCompactAndExpandedHelp(t *testing.T) {
	h := NewHeader(HeaderOptions{AppName: "Nu", Version: "dev", PaddingX: 1})

	compact := strings.Join(h.Render(80), "\n")
	if !strings.Contains(compact, "Nu vdev") {
		t.Fatalf("compact header = %q, want logo", compact)
	}
	if !strings.Contains(compact, "ctrl+o") {
		t.Fatalf("compact header = %q, want compact help", compact)
	}

	h.SetExpanded(true)
	expanded := strings.Join(h.Render(80), "\n")
	if !strings.Contains(expanded, "ctrl+c twice") {
		t.Fatalf("expanded header = %q, want expanded help", expanded)
	}
}
