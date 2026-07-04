package box

import (
	"testing"

	"nu/internal/tui/components/text"
)

func TestBoxRenderPadsChild(t *testing.T) {
	b := New(Options{PaddingX: 1, PaddingY: 1})
	b.AddChild(text.New("x", text.Options{}))

	lines := b.Render(5)
	if len(lines) != 3 || lines[1] != " x   " {
		t.Fatalf("lines = %#v", lines)
	}
}
