package components

// Fill expands to the remaining terminal rows assigned by the engine.
type Fill struct{}

// NewFill creates a flexible blank component.
func NewFill() *Fill {
	return &Fill{}
}

// Render returns no fixed lines; the engine calls FillLines with remaining rows.
func (f *Fill) Render(width int) []string {
	return nil
}
