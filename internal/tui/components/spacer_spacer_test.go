package components

import "testing"

func TestSpacerSpacerRender(t *testing.T) {
	lines := NewSpacer(2).Render(3)
	if len(lines) != 2 || lines[0] != "   " || lines[1] != "   " {
		t.Fatalf("lines = %#v", lines)
	}
}
