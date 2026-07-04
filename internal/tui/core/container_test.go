package core

import "testing"

type staticComponent []string

func (s staticComponent) Render(width int) []string {
	return []string(s)
}

func TestContainerRendersChildrenInOrder(t *testing.T) {
	var c Container
	c.AddChild(staticComponent{"a"})
	c.AddChild(staticComponent{"b"})

	got := c.Render(10)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("Render = %#v", got)
	}
}
