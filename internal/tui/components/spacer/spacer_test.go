package spacer

import "testing"

func TestSpacerRender(t *testing.T) {
	lines := New(2).Render(3)
	if len(lines) != 2 || lines[0] != "   " || lines[1] != "   " {
		t.Fatalf("lines = %#v", lines)
	}
}
