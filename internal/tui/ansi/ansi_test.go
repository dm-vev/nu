package ansi

import "testing"

func TestWrapTextWrapsInsteadOfTruncating(t *testing.T) {
	lines := WrapText("alpha beta gamma delta", 12)
	if len(lines) < 2 {
		t.Fatalf("lines = %#v, want wrapped lines", lines)
	}
	if lines[0] != "alpha beta" || lines[1] != "gamma delta" {
		t.Fatalf("lines = %#v", lines)
	}
}

func TestVisibleWidthIgnoresANSI(t *testing.T) {
	got := VisibleWidth(Green + "abc" + Reset)
	if got != 3 {
		t.Fatalf("VisibleWidth = %d, want 3", got)
	}
}
