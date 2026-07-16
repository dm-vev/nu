package components

import (
	"testing"
)

func TestBoxBoxRenderPadsChild(t *testing.T) {
	b := NewBox(BoxOptions{PaddingX: 1, PaddingY: 1})
	b.AddChild(NewText("x", TextOptions{}))

	lines := b.Render(5)
	if len(lines) != 3 || lines[1] != " x   " {
		t.Fatalf("lines = %#v", lines)
	}
}
