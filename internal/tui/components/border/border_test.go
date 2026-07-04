package border

import "testing"

func TestBorderRender(t *testing.T) {
	got := New(nil).Render(4)
	if len(got) != 1 || got[0] != "────" {
		t.Fatalf("Render = %#v", got)
	}
}
