package fill

import "testing"

func TestFillRendersAssignedRows(t *testing.T) {
	f := New()

	if lines := f.Render(10); len(lines) != 0 {
		t.Fatalf("Render lines = %#v, want no fixed lines", lines)
	}
	lines := f.FillLines(4, 2)
	if len(lines) != 2 || lines[0] != "    " || lines[1] != "    " {
		t.Fatalf("FillLines = %#v, want two padded lines", lines)
	}
}
