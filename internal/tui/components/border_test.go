package components

import "testing"

func TestBorderBorderRender(t *testing.T) {
	got := NewBorder(nil).Render(4)
	if len(got) != 1 || got[0] != "────" {
		t.Fatalf("Render = %#v", got)
	}
}
