package core

import "testing"

type coreStaticComponent []string

func (s coreStaticComponent) Render(width int) []string {
	return []string(s)
}

func TestCoreContainerRendersChildrenInOrder(t *testing.T) {
	var c Container
	c.AddChild(coreStaticComponent{"a"})
	c.AddChild(coreStaticComponent{"b"})

	got := c.Render(10)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("Render = %#v", got)
	}
}
