package core

// Container renders children in order.
type Container struct {
	Children []Component
}

// AddChild appends a component.
func (c *Container) AddChild(component Component) {
	c.Children = append(c.Children, component)
}

// RemoveChild removes a component by identity.
func (c *Container) RemoveChild(component Component) {
	for i, child := range c.Children {
		if child != component {
			continue
		}
		c.Children = append(c.Children[:i], c.Children[i+1:]...)
		return
	}
}

// Clear removes all children.
func (c *Container) Clear() {
	c.Children = nil
}

// Invalidate clears child caches.
func (c *Container) Invalidate() {
	for _, child := range c.Children {
		if invalid, ok := child.(Invalidatable); ok {
			invalid.Invalidate()
		}
	}
}

// Render renders every child at the same width.
func (c *Container) Render(width int) []string {
	lines := []string{}
	for _, child := range c.Children {
		lines = append(lines, child.Render(width)...)
	}
	return lines
}
